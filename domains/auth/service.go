package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rovilay/ecommerce-service/utils"
)

type AuthService interface {
	ValidateJWT(ctx context.Context, token string) (string, error)
}

type authService struct {
	authSecret []byte
	cache      *redis.Client
	expiration time.Duration
}

func NewAuthService(r *redis.Client, jwtSecret string, tokenExpiration time.Duration) *authService {
	return &authService{
		authSecret: []byte(jwtSecret),
		cache:      r,
		expiration: tokenExpiration,
	}
}

func (a *authService) ValidateJWT(ctx context.Context, token string) (string, error) {
	// 1. Check Redis Cache
	userID, err := a.cache.Get(ctx, token).Result()
	if err == nil {
		return userID, nil
	} else if err != redis.Nil {
		return "", err
	}

	// 2. Cache Miss - Perform full validation
	userID, err = utils.ValidateJWT(token, a.authSecret)
	if err != nil {
		return "", err
	}

	// 3. If valid, store in Redis with expiration
	err = a.cache.Set(ctx, token, userID, a.expiration).Err()
	if err != nil {
		return "", err
	}
	return userID, nil
}
