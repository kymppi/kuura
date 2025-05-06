package users

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

type TokenInfoForService struct {
	AccessToken         string
	AccessTokenDuration time.Duration
	RefreshToken        string
	SessionId           string
}

func hashCodeHMAC(code string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(code))
	return base64.RawStdEncoding.EncodeToString(h.Sum(nil))
}

func (s *UserService) ExchangeCodeForTokens(ctx context.Context, code string) (*TokenInfoForService, error) {
	hashedCode := hashCodeHMAC(code, s.tokenCodeHashingSecret)

	data, err := s.db.UseTokenExchangeCode(ctx, hashedCode)
	if err != nil {
		return nil, err
	}

	accessTokenDuration, err := s.db.GetAccessTokenDurationUsingSessionId(ctx, data.SessionID)
	if err != nil {
		return nil, err
	}

	decryptedAccessToken, err := s.encryptor.Decrypt([]byte(data.EncryptedAccessToken), s.encryptionKey, data.EncryptionNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access token: %w", err)
	}

	decryptedRefreshToken, err := s.encryptor.Decrypt([]byte(data.EncryptedRefreshToken), s.encryptionKey, data.EncryptionNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	return &TokenInfoForService{
		AccessToken:         string(decryptedAccessToken),
		AccessTokenDuration: time.Duration(accessTokenDuration) * time.Second,
		RefreshToken:        string(decryptedRefreshToken),
		SessionId:           data.SessionID,
	}, nil
}

func (s *UserService) RefreshServiceAccessToken(ctx context.Context, sessionId, refreshToken string) (*TokenInfoForService, error) {
	return nil, errors.New("not implemented")
}
