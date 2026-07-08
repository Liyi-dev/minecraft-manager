package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var Client *goredis.Client

func Init(addr, password string) error {
	Client = goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: ping failed: %w", err)
	}
	return nil
}

func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// BlacklistToken adds a token to the blacklist with its remaining TTL.
func BlacklistToken(token string, ttl time.Duration) error {
	if Client == nil {
		return fmt.Errorf("redis: client not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return Client.Set(ctx, "blacklist:"+token, "1", ttl).Err()
}

// IsTokenBlacklisted checks if a token is in the blacklist.
func IsTokenBlacklisted(token string) bool {
	if Client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := Client.Get(ctx, "blacklist:"+token).Result()
	return err == nil
}
