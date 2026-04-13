package storage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/config"
)

type Client struct {
	mc     *minio.Client
	Bucket string
}

// New connects to MinIO and ensures the bucket exists.
func New(ctx context.Context, cfg *config.Config) (*Client, error) {
	mc, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio new: %w", err)
	}

	exists, err := mc.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("bucket exists: %w", err)
	}
	if !exists {
		if err := mc.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("make bucket: %w", err)
		}
		log.Info().Str("bucket", cfg.MinioBucket).Msg("created minio bucket")
	}
	log.Info().Str("bucket", cfg.MinioBucket).Msg("minio ready")
	return &Client{mc: mc, Bucket: cfg.MinioBucket}, nil
}

func (c *Client) PresignedPut(ctx context.Context, key string, ttl time.Duration) (*url.URL, error) {
	return c.mc.PresignedPutObject(ctx, c.Bucket, key, ttl)
}

func (c *Client) PresignedGet(ctx context.Context, key string, ttl time.Duration) (*url.URL, error) {
	return c.mc.PresignedGetObject(ctx, c.Bucket, key, ttl, url.Values{})
}

func (c *Client) Remove(ctx context.Context, key string) error {
	return c.mc.RemoveObject(ctx, c.Bucket, key, minio.RemoveObjectOptions{})
}
