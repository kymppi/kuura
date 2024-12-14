package models

import (
	"time"

	"github.com/google/uuid"
)

type AppService struct {
	Id          uuid.UUID // uuidv7
	JWTAudience string
	CreatedAt   time.Time
	ModifiedAt  time.Time
	Name        string
	Description string // optional
}
