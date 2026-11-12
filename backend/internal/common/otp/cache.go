package otp

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const maxOtpAttempts = 5

type Cache struct {
	redis     *redis.Client
	keyPrefix string
	ttl       time.Duration
	cooldown  time.Duration
}

func NewCache(redis *redis.Client, purpose string) *Cache {
	return &Cache{
		redis:     redis,
		keyPrefix: "otp:" + purpose + ":",
		ttl:       5 * time.Minute,
		cooldown:  time.Minute,
	}
}

func (c *Cache) Put(ctx context.Context, email string, otp string) error {
	pipe := c.redis.TxPipeline()
	pipe.Set(ctx, c.key(email), otp, c.ttl)
	pipe.Del(ctx, c.attemptsKey(email))
	_, err := pipe.Exec(ctx)
	return err
}

func (c *Cache) OnCooldown(ctx context.Context, email string) (bool, error) {
	remaining, err := c.redis.TTL(ctx, c.key(email)).Result()
	if err != nil {
		return false, err
	}
	if remaining <= 0 {
		return false, nil
	}

	elapsed := c.ttl - remaining
	return elapsed < c.cooldown, nil
}

func (c *Cache) Verify(ctx context.Context, email string, otp string) (bool, error) {
	stored, err := c.redis.Get(ctx, c.key(email)).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if stored == otp {
		return true, c.Delete(ctx, email)
	}

	// Burn the code after too many wrong guesses so a 6-digit OTP can't be brute-forced within its TTL.
	attempts, err := c.redis.Incr(ctx, c.attemptsKey(email)).Result()
	if err != nil {
		return false, err
	}
	if attempts == 1 {
		c.redis.Expire(ctx, c.attemptsKey(email), c.ttl)
	}
	if attempts >= maxOtpAttempts {
		return false, c.Delete(ctx, email)
	}
	return false, nil
}

func (c *Cache) Delete(ctx context.Context, email string) error {
	return c.redis.Del(ctx, c.key(email), c.attemptsKey(email)).Err()
}

func (c *Cache) key(email string) string {
	return c.keyPrefix + email
}

func (c *Cache) attemptsKey(email string) string {
	return c.keyPrefix + "attempts:" + email
}
