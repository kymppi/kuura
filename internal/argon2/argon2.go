package tokenhasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
	ErrGeneratingSalt      = errors.New("failed to generate salt")
	ErrDecodeHash          = errors.New("failed to decode hash")
)

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type TokenHasher struct {
	params Argon2Params
}

func NewTokenHasher(params Argon2Params) *TokenHasher {
	return &TokenHasher{
		params: params,
	}
}

func (t *TokenHasher) HashValue(value string) (hashedValue string, err error) {
	salt, err := t.generateRandomBytes(t.params.SaltLength)

	if err != nil {
		return "", ErrGeneratingSalt
	}

	hashed := argon2.IDKey([]byte(value), salt, t.params.Iterations, t.params.Memory, t.params.Parallelism, t.params.KeyLength)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hashed)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, t.params.Memory, t.params.Iterations, t.params.Parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

func (t *TokenHasher) CompareHashAndValue(hashedValue string, value string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := t.decodeHash(hashedValue)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrDecodeHash, err)
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(value), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

func (t *TokenHasher) decodeHash(encodedHash string) (p *Argon2Params, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}

	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &Argon2Params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)

	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])

	if err != nil {
		return nil, nil, nil, err
	}

	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])

	if err != nil {
		return nil, nil, nil, err
	}

	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func (t *TokenHasher) generateRandomBytes(length uint32) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)

	if err != nil {
		return []byte{}, err
	}

	return b, nil
}
