package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type TokenInfoForService struct {
	AccessToken         string
	AccessTokenDuration time.Duration
	RefreshToken        string
	SessionId           string
}

func (s *UserService) ExchangeCodeForTokens(ctx context.Context, serviceId uuid.UUID, code string) (*TokenInfoForService, error) {
	return nil, errors.New("not implemented")
}

func (s *UserService) RefreshServiceAccessToken(ctx context.Context, serviceId uuid.UUID, sessionId, refreshToken string) (*TokenInfoForService, error) {
	return nil, errors.New("not implemented")
}
