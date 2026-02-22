package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{client: client}
}

func (s *Service) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return s.client.client.Set(ctx, key, value, expiration).Err()
}

func (s *Service) Get(ctx context.Context, key string) (string, error) {
	return s.client.client.Get(ctx, key).Result()
}

func (s *Service) Delete(ctx context.Context, key string) error {
	return s.client.client.Del(ctx, key).Err()
}

func (s *Service) Exists(ctx context.Context, key string) (bool, error) {
	result, err := s.client.client.Exists(ctx, key).Result()
	return result > 0, err
}

func (s *Service) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return s.client.client.Expire(ctx, key, expiration).Err()
}

func (s *Service) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.Set(ctx, key, jsonData, expiration)
}

func (s *Service) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := s.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

func (s *Service) StoreRefreshToken(ctx context.Context, userID, token string) error {
	key := fmt.Sprintf("refresh_token:%s", userID)
	return s.Set(ctx, key, token, 7*24*time.Hour)
}

func (s *Service) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", userID)
	return s.Get(ctx, key)
}

func (s *Service) DeleteRefreshToken(ctx context.Context, userID string) error {
	key := fmt.Sprintf("refresh_token:%s", userID)
	return s.Delete(ctx, key)
}

func (s *Service) StoreEmailConfirmation(ctx context.Context, userID, email, token string) error {
	userKey := fmt.Sprintf("email_confirm:user:%s", userID)
	tokenKey := fmt.Sprintf("email_confirm:token:%s", token)

	pipe := s.client.client.TxPipeline()
	pipe.Set(ctx, userKey, token, 24*time.Hour)
	pipe.Set(ctx, tokenKey, userID, 24*time.Hour)

	_, err := pipe.Exec(ctx)
	return err
}

func (s *Service) GetEmailConfirmationUserID(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("email_confirm:token:%s", token)
	return s.Get(ctx, key)
}

func (s *Service) DeleteEmailConfirmation(ctx context.Context, userID, token string) error {
	userKey := fmt.Sprintf("email_confirm:user:%s", userID)
	tokenKey := fmt.Sprintf("email_confirm:token:%s", token)

	pipe := s.client.client.TxPipeline()
	pipe.Del(ctx, userKey)
	pipe.Del(ctx, tokenKey)

	_, err := pipe.Exec(ctx)
	return err
}

func (s *Service) HealthCheck(ctx context.Context) error {
	return s.client.HealthCheck(ctx)
}
