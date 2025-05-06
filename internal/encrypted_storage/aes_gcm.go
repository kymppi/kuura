package encrypted_storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// SymmetricKeyEncryptor provides methods for encrypting and decrypting symmetric keys
// using AES-GCM (Galois/Counter Mode), which provides authenticated encryption
type SymmetricKeyEncryptor struct{}

// Encrypt encrypts the given data using the provided key and generates a random IV
//
// Parameters:
// - data: The data to be encrypted
// - key: The encryption key (must be 32 bytes/256 bits long)
//
// Returns:
// - ciphertext: The encrypted data
// - nonce: The Initialization Vector (IV) used for encryption
// - error: Any error encountered during encryption
func (e *SymmetricKeyEncryptor) Encrypt(data, key []byte) ([]byte, []byte, error) {
	// Validate key length - AES-256 requires a 32-byte key
	if len(key) != 32 {
		return nil, nil, errors.New("key must be 32 bytes long")
	}

	// Create a new AES cipher block using the provided key
	// This prepares the encryption algorithm with the specific key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	// Create a Galois/Counter Mode (GCM) block cipher mode
	// GCM provides authenticated encryption with built-in integrity checking
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	// Generate a Nonce (Number used ONCE) / Initialization Vector (IV)
	// - Critical for preventing replay attacks
	// - Must be unique for each encryption with the same key
	// - Typically random and the same length as the block cipher's nonce size
	// - In AES-GCM, this is usually 12 bytes (96 bits)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	// Encrypt the data:
	// - 'Seal' method combines encryption and authentication
	// - First nil argument means no additional data is prepended
	// - nonce is used as the initialization vector
	// - nil at the end means no additional authenticated data (AAD)
	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Return the encrypted data and the nonce (IV)
	return ciphertext, nonce, nil
}

// Decrypt decrypts the given data using the provided key and IV
//
// Parameters:
// - ciphertext: The encrypted data to be decrypted
// - key: The decryption key (must be 32 bytes/256 bits long)
// - nonce: The Initialization Vector (IV) used during encryption
//
// Returns:
// - plaintext: The decrypted original data
// - error: Any error encountered during decryption
func (e *SymmetricKeyEncryptor) Decrypt(ciphertext, key, nonce []byte) ([]byte, error) {
	// Validate key length - AES-256 requires a 32-byte key
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes long")
	}

	// Recreate the AES cipher block using the provided key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Recreate the GCM block cipher mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Verify that the nonce (IV) is the correct size
	// This prevents using an incorrectly sized initialization vector
	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("incorrect nonce size")
	}

	// Decrypt the data:
	// - 'Open' method verifies authentication and decrypts
	// - nil means no additional data was authenticated
	// - Returns the original plaintext if authentication succeeds
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// NewSymmetricKeyEncryptor creates a new SymmetricKeyEncryptor instance
// This is a convenience method for creating the encryptor
func NewSymmetricKeyEncryptor() *SymmetricKeyEncryptor {
	return &SymmetricKeyEncryptor{}
}
