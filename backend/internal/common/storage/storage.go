package storage

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	connectAttempts = 10
	retryDelay      = 2 * time.Second
)

type Client struct {
	client    *minio.Client
	bucket    string
	publicUrl string
}

func Connect(ctx context.Context) (*Client, error) {
	client, err := minio.New(os.Getenv("STORAGE_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("STORAGE_ACCESS_KEY"), os.Getenv("STORAGE_SECRET_KEY"), ""),
		Secure: os.Getenv("STORAGE_USE_SSL") == "true",
	})
	if err != nil {
		return nil, err
	}

	storage := &Client{
		client:    client,
		bucket:    os.Getenv("STORAGE_BUCKET"),
		publicUrl: strings.TrimRight(os.Getenv("STORAGE_PUBLIC_URL"), "/"),
	}

	for attempt := 1; ; attempt++ {
		err = storage.ensurePublicBucket(ctx)
		if err == nil || attempt >= connectAttempts {
			return storage, err
		}
		time.Sleep(retryDelay)
	}
}

func (c *Client) ensurePublicBucket(ctx context.Context) error {
	exists, err := c.client.BucketExists(ctx, c.bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := c.client.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}

	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, c.bucket)
	return c.client.SetBucketPolicy(ctx, c.bucket, policy)
}

func (c *Client) Upload(ctx context.Context, key string, content []byte, contentType string) (string, error) {
	_, err := c.client.PutObject(ctx, c.bucket, key, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{
		ContentType:  contentType,
		CacheControl: "public, max-age=31536000, immutable",
	})
	if err != nil {
		return "", err
	}
	return c.publicUrl + "/" + key, nil
}

func (c *Client) DeleteByUrl(ctx context.Context, url string) error {
	key, ok := strings.CutPrefix(url, c.publicUrl+"/")
	if !ok || key == "" {
		return nil
	}
	return c.client.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
}
