package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"lockerin-backend/internal/config"
	"lockerin-backend/internal/domain"

	"github.com/google/uuid"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type SupabaseAuthUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type SupabaseAuthResponse struct {
	AccessToken  string           `json:"access_token"`
	TokenType    string           `json:"token_type"`
	ExpiresIn    int64            `json:"expires_in"`
	RefreshToken string           `json:"refresh_token"`
	User         SupabaseAuthUser `json:"user"`
}

// SignUp registers a user via Supabase Auth Admin/Public REST API
func (c *Client) SignUp(ctx context.Context, email, password string) (*uuid.UUID, error) {
	if c.cfg.SupabaseURL == "" || c.cfg.SupabaseAnonKey == "" {
		// Fallback for standalone/mock mode
		id := uuid.New()
		return &id, nil
	}

	url := fmt.Sprintf("%s/auth/v1/signup", c.cfg.SupabaseURL)
	payload, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.cfg.SupabaseAnonKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("supabase auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("supabase signup error (%d): %v", resp.StatusCode, errResp["msg"])
	}

	var authResp SupabaseAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(authResp.User.ID)
	if err != nil {
		return nil, err
	}

	return &uid, nil
}

// SignIn verifies credentials with Supabase Auth
func (c *Client) SignIn(ctx context.Context, email, password string) (*SupabaseAuthResponse, error) {
	if c.cfg.SupabaseURL == "" || c.cfg.SupabaseAnonKey == "" {
		// Standalone / offline fallback
		return nil, domain.ErrNotFound
	}

	url := fmt.Sprintf("%s/auth/v1/token?grant_type=password", c.cfg.SupabaseURL)
	payload, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.cfg.SupabaseAnonKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("supabase sign-in failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.ErrUnauthorized
	}

	var authResp SupabaseAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, err
	}

	return &authResp, nil
}
