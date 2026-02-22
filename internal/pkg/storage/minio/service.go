package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{client}
}

func (s *Service) EnsureBucket(ctx context.Context, bucketName string) error {
	exists, err := s.client.GetClient().BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.GetClient().MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

func (s *Service) UploadFile(ctx context.Context, bucketName, objName string, file io.Reader, size int64) error {
	_, err := s.client.GetClient().PutObject(ctx, bucketName, objName, file, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *Service) DownloadFile(ctx context.Context, bucketName, objName string) (*minio.Object, error) {
	object, err := s.client.GetClient().GetObject(ctx, bucketName, objName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return object, nil
}

func (s *Service) GetFileURL(bucketName, objName string) string {
	return fmt.Sprintf("http://%s/%s/%s", s.client.config.Endpoint, bucketName, objName)
}

func (s *Service) DeleteFile(ctx context.Context, bucketName, objName string) error {
	err := s.client.GetClient().RemoveObject(ctx, bucketName, objName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *Service) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	buckets, err := s.client.GetClient().ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	return buckets, nil
}

func (s *Service) CreatePresignedURL(ctx context.Context, bucketName, objName string, expirySec int) (string, error) {
	url, err := s.client.GetClient().PresignedGetObject(ctx, bucketName, objName, time.Duration(expirySec)*time.Second, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create presidned URL: %w", err)
	}

	return url.String(), nil
}

func (s *Service) FileExists(ctx context.Context, bucketName, objName string) (bool, error) {
	_, err := s.client.GetClient().StatObject(ctx, bucketName, objName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return true, nil
}

func (s *Service) GenerateDownloadURL(ctx context.Context, bucketName, objName, filename string) (string, error) {
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	url, err := s.client.GetClient().PresignedGetObject(ctx, bucketName, objName, 15*time.Minute, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to create download URL: %w", err)
	}

	return url.String(), nil
}

func (s *Service) GenerateUploadURL(ctx context.Context, bucketName, objName string) (string, error) {
	url, err := s.client.GetClient().PresignedPutObject(ctx, bucketName, objName, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to create upload URL: %w", err)
	}

	return url.String(), nil
}

func (s *Service) ListObjects(ctx context.Context, bucketName, prefix string) ([]minio.ObjectInfo, error) {
	var objects []minio.ObjectInfo

	for obj := range s.client.GetClient().ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", obj.Err)
		}
		objects = append(objects, obj)
	}

	return objects, nil
}

func (s *Service) GetFileInfo(ctx context.Context, bucketName, objName string) (minio.ObjectInfo, error) {
	info, err := s.client.GetClient().StatObject(ctx, bucketName, objName, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("failed to get file info: %w", err)
	}

	return info, nil
}
