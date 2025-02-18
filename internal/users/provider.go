package users

import (
	"log/slog"

	tokenhasher "github.com/kymppi/kuura/internal/argon2"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/jwks"
)

type UserService struct {
	logger      *slog.Logger
	db          *db_gen.Queries
	tokenhasher *tokenhasher.TokenHasher
	jwtIssuer   string
	jwkManager  *jwks.JWKManager
}

func NewUserService(logger *slog.Logger, db *db_gen.Queries, jwtIssuer string, jwkManager *jwks.JWKManager) *UserService {
	return &UserService{
		logger: logger,
		db:     db,
		tokenhasher: tokenhasher.NewTokenHasher(tokenhasher.Argon2Params{
			Memory:      64 * 1024,
			Iterations:  3,
			Parallelism: 2,
			SaltLength:  16,
			KeyLength:   32,
		}),
		jwtIssuer:  jwtIssuer,
		jwkManager: jwkManager,
	}
}
