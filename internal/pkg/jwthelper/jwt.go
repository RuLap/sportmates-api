package jwthelper

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTHelper struct {
	secret []byte
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func NewJwtHelper(secret string) (*JWTHelper, error) {
	if secret == "" {
		return nil, errors.New("empty JWT secret")
	}
	return &JWTHelper{secret: []byte(secret)}, nil
}

func (h *JWTHelper) GenerateJWT(userID, email, tokenType string, expiresIn time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiresIn)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sportmates-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.secret)
}

func (h *JWTHelper) GenerateTokenPair(userID, email string) (*TokenPair, error) {
	accessToken, err := h.GenerateJWT(userID, email, "access", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshToken, err := h.GenerateJWT(userID, email, "refresh", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(15 * time.Minute / time.Second),
	}, nil
}

func (h *JWTHelper) GenerateDefaultToken(userID, email string) (string, error) {
	return h.GenerateJWT(userID, email, "access", 24*time.Hour)
}

func (h *JWTHelper) ParseJWT(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("empty token string")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return h.secret, nil
		},
	)

	if err != nil {
		return nil, errors.New("invalid token")
	}

	if token == nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (h *JWTHelper) ValidateAccessToken(tokenString string) (bool, error) {
	claims, err := h.ParseJWT(tokenString)
	if err != nil {
		return false, err
	}

	if claims.Type != "access" {
		return false, errors.New("not an access token")
	}

	return true, nil
}

func (h *JWTHelper) ValidateRefreshToken(tokenString string) (bool, error) {
	claims, err := h.ParseJWT(tokenString)
	if err != nil {
		return false, err
	}

	if claims.Type != "refresh" {
		return false, errors.New("not a refresh token")
	}

	return true, nil
}

func (h *JWTHelper) ValidateToken(tokenString string) (bool, error) {
	return h.ValidateAccessToken(tokenString)
}
