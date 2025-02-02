package srp

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"
)

type SRPClient struct {
	Prime     *big.Int
	Generator *big.Int
	Key       *big.Int
}

func NewSRPClient(prime, generator string, key *big.Int) (*SRPClient, error) {
	p, ok := new(big.Int).SetString(prime, 16)
	if !ok {
		return nil, errors.New("invalid prime")
	}
	g, ok := new(big.Int).SetString(generator, 16)
	if !ok {
		return nil, errors.New("invalid generator")
	}

	return &SRPClient{
		Prime:     p,
		Generator: g,
		Key:       key,
	}, nil
}

func (c *SRPClient) GenerateSalt() (*big.Int, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(bytes), nil
}

func (c *SRPClient) GeneratePasswordHash(salt *big.Int, identity, password string) (*big.Int, error) {
	h := sha256.New()
	h.Write([]byte(identity + ":" + password))
	hash1 := h.Sum(nil)

	h.Reset()
	h.Write(salt.Bytes())
	h.Write(hash1)
	hash2 := h.Sum(nil)

	return new(big.Int).SetBytes(hash2), nil
}

func (c *SRPClient) GenerateVerifier(x *big.Int) *big.Int {
	return new(big.Int).Exp(c.Generator, x, c.Prime)
}

func (c *SRPClient) GeneratePublic(secret *big.Int) *big.Int {
	return new(big.Int).Exp(c.Generator, secret, c.Prime)
}

func (c *SRPClient) GeneratePreMasterSecret(a, B, x, u *big.Int) (*big.Int, error) {
	if new(big.Int).Mod(B, c.Prime).Cmp(big.NewInt(0)) == 0 {
		return nil, errors.New("server may return an invalid public ephemeral")
	}

	k := c.Key
	g := c.Generator
	N := c.Prime

	gx := new(big.Int).Exp(g, x, N)
	kTimesGx := new(big.Int).Mul(k, gx)
	B2 := new(big.Int).Sub(B, kTimesGx)
	if B2.Cmp(big.NewInt(0)) < 0 {
		B2 = new(big.Int).Mod(new(big.Int).Neg(B2), N)
	}

	exp := new(big.Int).Add(a, new(big.Int).Mul(u, x))
	S := new(big.Int).Exp(B2, exp, N)

	return S, nil
}

func (c *SRPClient) Register(identity, password string) (salt *big.Int, verifier *big.Int, err error) {
	salt, err = c.GenerateSalt()
	if err != nil {
		return nil, nil, err
	}

	x, err := c.GeneratePasswordHash(salt, identity, password)
	if err != nil {
		return nil, nil, err
	}

	verifier = c.GenerateVerifier(x)
	return salt, verifier, nil
}
