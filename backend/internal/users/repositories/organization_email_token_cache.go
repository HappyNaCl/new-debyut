package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type VerifiedOrganizationEmail struct {
	UserId string
	Email  string
}

type OrganizationEmailTokenCache struct {
	redis     *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewOrganizationEmailTokenCache(redis *redis.Client) *OrganizationEmailTokenCache {
	return &OrganizationEmailTokenCache{
		redis:     redis,
		keyPrefix: "organization-email-token:",
		ttl:       30 * time.Minute,
	}
}

func (c *OrganizationEmailTokenCache) Put(ctx context.Context, token string, verified VerifiedOrganizationEmail) error {
	value, err := json.Marshal(verified)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, c.key(token), value, c.ttl).Err()
}

func (c *OrganizationEmailTokenCache) Get(ctx context.Context, token string) (*VerifiedOrganizationEmail, error) {
	value, err := c.redis.Get(ctx, c.key(token)).Bytes()
	if err != nil {
		return nil, err
	}

	var verified VerifiedOrganizationEmail
	if err := json.Unmarshal(value, &verified); err != nil {
		return nil, err
	}
	return &verified, nil
}

func (c *OrganizationEmailTokenCache) Delete(ctx context.Context, token string) error {
	return c.redis.Del(ctx, c.key(token)).Err()
}

func (c *OrganizationEmailTokenCache) key(token string) string {
	return c.keyPrefix + token
}
