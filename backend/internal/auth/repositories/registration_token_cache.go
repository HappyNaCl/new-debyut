package repositories

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RegistrationTokenCache struct {
	redis     *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewRegistrationTokenCache(redis *redis.Client) *RegistrationTokenCache {
	return &RegistrationTokenCache{
		redis:     redis,
		keyPrefix: "registration-token:",
		ttl:       30 * time.Minute,
	}
}

func (c *RegistrationTokenCache) Put(ctx context.Context, token string, email string) error {
	return c.redis.Set(ctx, c.key(token), email, c.ttl).Err()
}

func (c *RegistrationTokenCache) Get(ctx context.Context, token string) (string, error) {
	return c.redis.Get(ctx, c.key(token)).Result()
}

func (c *RegistrationTokenCache) Delete(ctx context.Context, token string) error {
	return c.redis.Del(ctx, c.key(token)).Err()
}

func (c *RegistrationTokenCache) key(token string) string {
	return c.keyPrefix + token
}
