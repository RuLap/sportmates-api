package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/RuLap/sportmates-api/internal/pkg/errors"
	"github.com/RuLap/sportmates-api/internal/pkg/events"
	"github.com/RuLap/sportmates-api/internal/pkg/jwthelper"
	"github.com/RuLap/sportmates-api/internal/pkg/rabbitmq"
	"github.com/RuLap/sportmates-api/internal/pkg/redis"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, userID string) error

	SendConfirmationLink(ctx context.Context, req *SendConfirmationEmailRequest, userID string) error
	ConfirmEmail(ctx context.Context, token string) error
	IsEmailConfirmed(ctx context.Context, userID uuid.UUID) (bool, error)
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type service struct {
	log       *slog.Logger
	jwtHelper *jwthelper.JWTHelper
	redis     *redis.Service
	rabbitmq  *rabbitmq.Service
	repo      Repository
}

func NewService(
	log *slog.Logger,
	jwtHelper *jwthelper.JWTHelper,
	redis *redis.Service,
	rabbitmq *rabbitmq.Service,
	repo Repository,
) Service {
	return &service{
		log:       log,
		jwtHelper: jwtHelper,
		redis:     redis,
		rabbitmq:  rabbitmq,
		repo:      repo,
	}
}

func (s *service) SendConfirmationLink(ctx context.Context, req *SendConfirmationEmailRequest, userID string) error {
	if userID == "" || req.Email == "" {
		return fmt.Errorf("userID и email обязательны")
	}

	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		s.log.Error("failed to generate token", "error", err)
		return fmt.Errorf("не удалось сгенерировать токен")
	}
	token := hex.EncodeToString(rawToken)

	err := s.redis.StoreEmailConfirmation(ctx, userID, req.Email, token)
	if err != nil {
		s.log.Error("failed to store token in redis", "error", err, "user_id", userID)
		return fmt.Errorf("не удалось сохранить токен")
	}

	confirmationURL := fmt.Sprintf("https://sportmates.ru/confirm?token=%s", token)

	if s.rabbitmq != nil {
		event := events.EmailEvent{
			To:       req.Email,
			Template: "email_confirmation",
			Subject:  "Подтвердите ваш email",
			Data: map[string]interface{}{
				"confirmation_url": confirmationURL,
				"user_email":       req.Email,
			},
		}

		if err := s.rabbitmq.PublishEmail(event); err != nil {
			s.log.Error("failed to publish email event", "error", err)
		}
	} else {
		s.log.Warn("event service not available - email not sent")
	}

	s.log.Info("confirmation link generated and sent", "email", req.Email, "user_id", userID)
	return nil
}

func (s *service) ConfirmEmail(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("токен обязателен")
	}

	userID, err := s.redis.GetEmailConfirmationUserID(ctx, token)
	if err != nil {
		s.log.Warn("invalid or expired confirmation token", "token", token, "error", err)
		return fmt.Errorf("неверная или устаревшая ссылка подтверждения")
	}

	if err := s.repo.MakeEmailConfirmed(ctx, userID); err != nil {
		s.log.Error("failed to confirm email in database", "error", err, "user_id", userID)
		return fmt.Errorf("не удалось подтвердить email")
	}

	if err := s.redis.DeleteEmailConfirmation(ctx, userID, token); err != nil {
		s.log.Warn("failed to delete used token", "token", token, "error", err)
	}

	s.log.Info("email confirmed successfully", "user_id", userID)
	return nil
}

func (s *service) IsEmailConfirmed(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf(errors.ErrFailedToLoadData)
	}

	return user.EmailConfirmed, nil
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to hash password", "error", err)
		return nil, fmt.Errorf("произошла ошибка")
	}

	hashedPasswordStr := string(hashedPassword)

	user := RegisterRequestToUser(&req, hashedPasswordStr)

	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		s.log.Error("failed to create user", "error", err, "email", user.Email)
		return nil, err
	}

	tokenPair, err := s.jwtHelper.GenerateTokenPair(*userID, req.Email)
	if err != nil {
		s.log.Error("failed to generate JWT tokens", "error", err)
		return nil, fmt.Errorf("произошла ошибка")
	}

	err = s.storeRefreshToken(ctx, *userID, tokenPair.RefreshToken)
	if err != nil {
		s.log.Error("failed to store refresh token", "error", err, "user_id", *userID)
		return nil, fmt.Errorf("произошла ошибка")
	}

	s.log.Info("user registered successfully", "user_id", *userID, "email", req.Email)

	return &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		UserID:       *userID,
		Email:        req.Email,
	}, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetByEmailProvider(ctx, req.Email, LocalProvider)
	if err != nil {
		s.log.Warn("user not found", "email", req.Email)
		return nil, fmt.Errorf("неверный email или пароль")
	}

	passwordHash, err := s.repo.GetPasswordHashByEmail(ctx, req.Email)
	if err != nil {
		s.log.Warn("failed to get password hash", "email", req.Email, "error", err)
		return nil, fmt.Errorf("неверный email или пароль")
	}

	if req.Password == "" {
		s.log.Error("user entered empty password", "email", req.Email)
		return nil, fmt.Errorf("неверный email или пароль")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*passwordHash), []byte(req.Password)); err != nil {
		s.log.Error("user entered invalid password", "email", req.Email)
		return nil, fmt.Errorf("неверный email или пароль")
	}

	tokenPair, err := s.jwtHelper.GenerateTokenPair(user.ID.String(), user.Email)
	if err != nil {
		s.log.Error("failed to generate JWT tokens", "error", err)
		return nil, fmt.Errorf("произошла ошибка")
	}

	err = s.storeRefreshToken(ctx, user.ID.String(), tokenPair.RefreshToken)
	if err != nil {
		s.log.Error("failed to store refresh token", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("произошла ошибка")
	}

	s.log.Info("user logged in successfully", "user_id", user.ID, "email", req.Email)

	return &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		UserID:       user.ID.String(),
		Email:        user.Email,
	}, nil
}

func (s *service) RefreshTokens(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token обязателен")
	}

	claims, err := s.jwtHelper.ParseJWT(refreshToken)
	if err != nil {
		s.log.Warn("invalid refresh token format", "error", err)
		return nil, fmt.Errorf("неверный refresh token")
	}

	if claims.Type != "refresh" {
		s.log.Warn("attempt to use non-refresh token for refresh", "token_type", claims.Type)
		return nil, fmt.Errorf("неверный тип токена")
	}

	storedToken, err := s.redis.GetRefreshToken(ctx, claims.UserID)
	if err != nil {
		s.log.Warn("refresh token not found in storage", "user_id", claims.UserID, "error", err)
		return nil, fmt.Errorf("refresh token не найден или истек")
	}

	if storedToken != refreshToken {
		s.log.Warn("refresh token mismatch", "user_id", claims.UserID)
		return nil, fmt.Errorf("неверный refresh token")
	}

	newTokenPair, err := s.jwtHelper.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		s.log.Error("failed to generate new token pair", "error", err, "user_id", claims.UserID)
		return nil, fmt.Errorf("произошла ошибка")
	}

	err = s.storeRefreshToken(ctx, claims.UserID, newTokenPair.RefreshToken)
	if err != nil {
		s.log.Error("failed to store new refresh token", "error", err, "user_id", claims.UserID)
		return nil, fmt.Errorf("произошла ошибка")
	}

	s.log.Info("tokens refreshed successfully", "user_id", claims.UserID)

	return &AuthResponse{
		AccessToken:  newTokenPair.AccessToken,
		RefreshToken: newTokenPair.RefreshToken,
		ExpiresIn:    newTokenPair.ExpiresIn,
		UserID:       claims.UserID,
		Email:        claims.Email,
	}, nil
}

func (s *service) Logout(ctx context.Context, userID string) error {
	err := s.redis.DeleteRefreshToken(ctx, userID)
	if err != nil {
		s.log.Error("failed to delete refresh token", "error", err, "user_id", userID)
		return fmt.Errorf("не удалось выполнить выход")
	}

	s.log.Info("user logged out successfully", "user_id", userID)
	return nil
}

func (s *service) ValidateToken(token string) (bool, error) {
	valid, err := s.jwtHelper.ValidateToken(token)
	if err != nil {
		s.log.Warn("token validation failed", "error", err)
		return false, fmt.Errorf("неверный токен")
	}
	return valid, nil
}

func (s *service) storeRefreshToken(ctx context.Context, userID, refreshToken string) error {
	return s.redis.StoreRefreshToken(ctx, userID, refreshToken)
}

func (s *service) getRefreshTokenKey(userID string) string {
	return fmt.Sprintf("refresh_token:%s", userID)
}
