package srp

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateRandomKey(prime *big.Int) (*big.Int, error) {
	// generate a random number in the range [1, prime-1].
	max := new(big.Int).Sub(prime, big.NewInt(1))
	key, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %v", err)
	}

	return key, nil
}
