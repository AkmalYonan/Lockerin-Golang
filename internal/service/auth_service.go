package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"lockerin-backend/internal/auth"
	"lockerin-backend/internal/domain"
	"lockerin-backend/internal/integration/supabase"
	"lockerin-backend/internal/repository"

	"github.com/google/uuid"
)

type AuthService struct {
	profileRepo    repository.ProfileRepository
	jwtManager     *auth.JWTManager
	supabaseClient *supabase.Client
}

func NewAuthService(profileRepo repository.ProfileRepository, jwtManager *auth.JWTManager, supabaseClient *supabase.Client) *AuthService {
	return &AuthService{
		profileRepo:    profileRepo,
		jwtManager:     jwtManager,
		supabaseClient: supabaseClient,
	}
}

func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, domain.NewAppError(400, "VALIDATION_ERROR", "Email, password, and name are required", nil)
	}

	// Check if already registered
	existing, _ := s.profileRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, domain.NewAppError(409, "USER_EXISTS", "User with this email already exists", nil)
	}

	// Try Supabase Auth Sign Up if enabled
	var userID uuid.UUID
	supabaseID, err := s.supabaseClient.SignUp(ctx, req.Email, req.Password)
	if err == nil && supabaseID != nil {
		userID = *supabaseID
	} else {
		userID = uuid.New()
	}

	profile := &domain.Profile{
		ID:        userID,
		Email:     req.Email,
		Username:  req.Username,
		Name:      req.Name,
		Phone:     req.Phone,
		Role:      domain.RoleUser,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.profileRepo.Create(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	token, exp, err := s.jwtManager.GenerateToken(profile)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: token,
		ExpiresIn:   exp,
		TokenType:   "Bearer",
		Profile:     profile,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	if req.EmailOrUsername == "" || req.Password == "" {
		return nil, domain.NewAppError(400, "VALIDATION_ERROR", "Email and password are required", nil)
	}

	var profile *domain.Profile
	var err error

	if strings.Contains(req.EmailOrUsername, "@") {
		profile, err = s.profileRepo.GetByEmail(ctx, req.EmailOrUsername)
	} else {
		profile, err = s.profileRepo.GetByUsername(ctx, req.EmailOrUsername)
	}

	if err != nil || profile == nil {
		return nil, domain.NewAppError(401, "INVALID_CREDENTIALS", "Invalid email/username or password", nil)
	}

	// Check status
	if profile.Status != "active" {
		return nil, domain.NewAppError(403, "ACCOUNT_SUSPENDED", "Your account is not active", nil)
	}

	token, exp, err := s.jwtManager.GenerateToken(profile)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: token,
		ExpiresIn:   exp,
		TokenType:   "Bearer",
		Profile:     profile,
	}, nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	return s.profileRepo.GetByID(ctx, userID)
}
