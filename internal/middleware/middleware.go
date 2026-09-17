package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"lockerin-backend/internal/auth"
	"lockerin-backend/internal/domain"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserClaimKey contextKey = "user_claims"
)

// RequestID Middleware
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			bytes := make([]byte, 8)
			_, _ = rand.Read(bytes)
			reqID = fmt.Sprintf("req-%s", hex.EncodeToString(bytes))
		}
		w.Header().Set("X-Request-ID", reqID)
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if val, ok := ctx.Value(RequestIDKey).(string); ok {
		return val
	}
	return ""
}

// Logger Middleware
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		srw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(srw, r)

		reqID := GetRequestID(r.Context())
		duration := time.Since(start)
		log.Printf("[%s] %s %s %d (%v)", reqID, r.Method, r.URL.Path, srw.statusCode, duration)
	})
}

// Recoverer Middleware
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				reqID := GetRequestID(r.Context())
				log.Printf("[PANIC RECOVERED] [%s]: %v", reqID, rvr)
				domain.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred", nil, reqID)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Auth Middleware
func AuthMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := GetRequestID(r.Context())
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				domain.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header", nil, reqID)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				domain.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token format", nil, reqID)
				return
			}

			claims, err := jwtManager.VerifyToken(parts[1])
			if err != nil {
				domain.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token", nil, reqID)
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserClaims(ctx context.Context) *auth.Claims {
	if val, ok := ctx.Value(UserClaimKey).(*auth.Claims); ok {
		return val
	}
	return nil
}

// RequireRole Middleware (RBAC)
func RequireRole(roles ...domain.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := GetRequestID(r.Context())
			claims := GetUserClaims(r.Context())
			if claims == nil {
				domain.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
				return
			}

			allowed := false
			for _, role := range roles {
				if claims.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				domain.WriteError(w, http.StatusForbidden, "FORBIDDEN", fmt.Sprintf("Access denied: role '%s' insufficient", claims.Role), nil, reqID)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter
type RateLimiter struct {
	mu      sync.Mutex
	limits  map[string]*clientRate
	maxReq  int
	window  time.Duration
}

type clientRate struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(maxReq int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limits: make(map[string]*clientRate),
		maxReq: maxReq,
		window: window,
	}
	go rl.cleanupRoutine()
	return rl
}

func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.window)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, v := range rl.limits {
			if now.After(v.resetTime) {
				delete(rl.limits, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Limit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = strings.Split(forwarded, ",")[0]
			}

			rl.mu.Lock()
			now := time.Now()
			cr, exists := rl.limits[ip]
			if !exists || now.After(cr.resetTime) {
				rl.limits[ip] = &clientRate{count: 1, resetTime: now.Add(rl.window)}
				rl.mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if cr.count >= rl.maxReq {
				rl.mu.Unlock()
				reqID := GetRequestID(r.Context())
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(cr.resetTime).Seconds())))
				domain.WriteError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests. Please wait.", nil, reqID)
				return
			}

			cr.count++
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}
