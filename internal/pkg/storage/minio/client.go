package minio

import (
	"context"
	"fmt"

	"github.com/RuLap/sportmates-api/internal/pkg/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client *minio.Client
	config *config.MinioConfig
}

func New(cfg *config.MinioConfig) (*Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	client := &Client{
		client: minioClient,
		config: cfg,
	}

	if err := client.HealthCheck(); err != nil {
		return nil, fmt.Errorf("minio health check failed: %s", cfg.Endpoint)
	}

	return client, nil
}

func (c *Client) HealthCheck() error {
	ctx := context.Background()
	_, err := c.client.ListBuckets(ctx)
	return err
}

func (c *Client) GetClient() *minio.Client {
	return c.client
}
