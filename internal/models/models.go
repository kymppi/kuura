package models

import (
	"time"

	"github.com/google/uuid"
)

type AppService struct {
	Id          uuid.UUID `json:"id" yaml:"id"` // uuidv7
	JWTAudience string    `json:"jwt_audience" yaml:"jwt_audience"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	ModifiedAt  time.Time `json:"modified_at" yaml:"modified_at"`
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description" yaml:"description"` // optional
}
