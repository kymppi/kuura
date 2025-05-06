package users

import (
	"log/slog"

	tokenhasher "github.com/kymppi/kuura/internal/argon2"
	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/encrypted_storage"
	"github.com/kymppi/kuura/internal/jwks"
	"github.com/kymppi/kuura/internal/services"
)

type UserService struct {
	logger      *slog.Logger
	db          *db_gen.Queries
	tokenhasher *tokenhasher.TokenHasher
	jwtIssuer   string
	jwkManager  *jwks.JWKManager
	services    *services.ServiceManager

	encryptor              *encrypted_storage.SymmetricKeyEncryptor
	encryptionKey          []byte
	tokenCodeHashingSecret []byte
}

func NewUserService(logger *slog.Logger, db *db_gen.Queries, jwtIssuer string, jwkManager *jwks.JWKManager, services *services.ServiceManager, encryptionKey []byte, tokenCodeHashingSecret []byte) *UserService {
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
		jwtIssuer:              jwtIssuer,
		jwkManager:             jwkManager,
		services:               services,
		encryptor:              encrypted_storage.NewSymmetricKeyEncryptor(),
		encryptionKey:          encryptionKey,
		tokenCodeHashingSecret: tokenCodeHashingSecret,
	}
}
