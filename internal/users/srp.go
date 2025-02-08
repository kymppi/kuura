package users

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kymppi/kuura/internal/db_gen"
)

func (s *UserService) SRPVerify(ctx context.Context, identity string, premaster string) error {
	uid, err := s.db.GetUserIDFromUsername(ctx, identity)
	if err != nil {
		return err
	}

	storedPremaster, err := s.db.GetAndDeletePremaster(ctx, uid)
	if err != nil {
		return err
	}

	if storedPremaster != premaster {
		return errors.New("premasters don't match")
	}

	return nil
}

func (s *UserService) SRPChallenge(ctx context.Context, identity string, A *big.Int) (salt string, B *big.Int, err error) {
	N, ok := new(big.Int).SetString(s.srp.PrimeHex, 16)
	if !ok {
		return "", nil, errors.New("invalid prime")
	}
	g, ok := new(big.Int).SetString(s.srp.Generator, 16)
	if !ok {
		return "", nil, errors.New("invalid generator")
	}

	row, err := s.db.GetUserSaltAndVerifier(ctx, identity)
	if err != nil {
		return "", nil, err
	}

	salt = row.Salt
	v := new(big.Int).SetBytes([]byte(row.Verifier))

	b, B := calculateBandB(v, N, g)

	/*
		1. s, v = lookup(I)
		2. gen b and B
		3. calculate premaster and store in db with 30s TTL
		4. return s and B
	*/

	premaster, err := calculatePremaster(b, v, N, A, B)
	if err != nil {
		return "", nil, fmt.Errorf("failed to calculate premaster: %w", err)
	}

	err = s.db.StorePremaster(ctx, db_gen.StorePremasterParams{
		ID:   row.ID,
		Data: premaster,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(30 * time.Second),
			Valid: true,
		},
	})
	if err != nil {
		return "", nil, err
	}

	s.logger.Info("Stored", slog.String("premaster", premaster))

	return salt, B, nil
}

func calculateBandB(v *big.Int, N *big.Int, g *big.Int) (*big.Int, *big.Int) {
	b := random()
	k := new(big.Int).SetBytes(sha256Hash(append(N.Bytes(), g.Bytes()...)))
	gb := new(big.Int).Exp(g, b, N)
	B := new(big.Int).Add(new(big.Int).Mul(k, v), gb)
	B.Mod(B, N)
	return b, B
}

func calculatePremaster(b *big.Int, v *big.Int, N *big.Int, A *big.Int, B *big.Int) (string, error) {
	// u = SHA256(PAD(A) | PAD(B))
	paddedA := pad(A.Bytes(), 64)
	paddedB := pad(B.Bytes(), 64)
	u := new(big.Int).SetBytes(sha256Hash(append(paddedA, paddedB...)))

	// (A * v^u) ^ b % N
	vU := new(big.Int).Exp(v, u, N)
	AvU := new(big.Int).Mul(A, vU)
	AvU.Mod(AvU, N)

	premaster := new(big.Int).Exp(AvU, b, N)
	return premaster.String(), nil
}

func random() *big.Int {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return new(big.Int).SetBytes(b)
}

func pad(b []byte, blockSize int) []byte {
	paddingSize := blockSize - (len(b) % blockSize) - 1
	padded := make([]byte, len(b)+paddingSize+1+8)
	copy(padded, b)
	padded[len(b)] = 0x80
	for i := len(b) + 1; i < len(padded)-8; i++ {
		padded[i] = 0x00
	}
	binary.BigEndian.PutUint64(padded[len(padded)-8:], uint64(len(b)*8))
	return padded
}

func sha256Hash(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	return h.Sum(nil)
}
