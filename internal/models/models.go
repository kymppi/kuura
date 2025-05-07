package models

import (
	"time"

	"github.com/google/uuid"
)

type AppService struct {
	Id                      uuid.UUID     `json:"id" yaml:"id"` // uuidv7
	JWTAudience             string        `json:"jwt_audience" yaml:"jwt_audience"`
	CreatedAt               time.Time     `json:"created_at" yaml:"created_at"`
	ModifiedAt              time.Time     `json:"modified_at" yaml:"modified_at"`
	Name                    string        `json:"name" yaml:"name"`
	Description             string        `json:"description" yaml:"description"` // optional
	ContactName             string        `json:"contact_name" yaml:"contact_name"`
	ContactEmail            string        `json:"contact_email" yaml:"contact_email"`
	LoginRedirect           string        `json:"login_redirect" yaml:"login_redirect"`
	AccessTokenDuration     time.Duration `json:"access_token_duration" yaml:"access_token_duration"`
	AccessTokenCookieDomain string        `json:"access_token_domain" yaml:"access_token_domain"`
	AccessTokenCookie       string        `json:"access_token_cookie" yaml:"access_token_cookie"`
}

type M2MRoleTemplate struct {
	Id    string   `json:"id" yaml:"id"`
	Roles []string `json:"roles" yaml:"roles"`
}

type User struct {
	Id          string
	Username    string
	LastLoginAt *time.Time
}

type UserSession struct {
	Id                  string
	UserId              string
	ServiceId           *uuid.UUID
	RefreshTokenHash    string
	ExpiresAt           time.Time
	CreatedAt           time.Time
	LastAuthenticatedAt time.Time
}
