package jwk_storage

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// Predefined storage key used for encryption/decryption
	testStorageKey = hexToBytes("00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff")

	// Predefined key to be encrypted
	testEncryptableData = hexToBytes("fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210")

	// Another predefined key for additional testing
	testAlternateKey = hexToBytes("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
)

func hexToBytes(hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return bytes
}

func TestSymmetricKeyEncryptor(t *testing.T) {
	encryptor := NewSymmetricKeyEncryptor()

	t.Run("Encrypt and Decrypt Successful", func(t *testing.T) {
		ciphertext, nonce, err := encryptor.Encrypt(testEncryptableData, testStorageKey)
		assert.NoError(t, err, "Encryption should succeed")
		assert.NotEmpty(t, ciphertext, "Ciphertext should not be empty")
		assert.NotEmpty(t, nonce, "Nonce should not be empty")

		decryptedKey, err := encryptor.Decrypt(ciphertext, testStorageKey, nonce)
		assert.NoError(t, err, "Decryption should succeed")

		assert.Equal(t, testEncryptableData, decryptedKey, "Decrypted key should match original")
	})

	t.Run("Decryption with Incorrect Storage Key", func(t *testing.T) {
		ciphertext, nonce, err := encryptor.Encrypt(testEncryptableData, testStorageKey)
		assert.NoError(t, err, "Encryption should succeed")

		_, err = encryptor.Decrypt(ciphertext, testAlternateKey, nonce)
		assert.Error(t, err, "Decryption with incorrect storage key should fail")
	})

	t.Run("Encryption and Decryption of Different Keys", func(t *testing.T) {
		ciphertext, nonce, err := encryptor.Encrypt(testAlternateKey, testStorageKey)
		assert.NoError(t, err, "Encryption should succeed")

		decryptedKey, err := encryptor.Decrypt(ciphertext, testStorageKey, nonce)
		assert.NoError(t, err, "Decryption should succeed")

		assert.Equal(t, testAlternateKey, decryptedKey, "Decrypted alternate key should match original")
	})

	t.Run("Encryption with Invalid Key Length", func(t *testing.T) {
		shortKey := make([]byte, 16)
		_, _, err := encryptor.Encrypt(testEncryptableData, shortKey)
		assert.Error(t, err, "Encryption with short key should fail")
		assert.Contains(t, err.Error(), "key must be 32 bytes long")
	})

	t.Run("Decryption with Invalid Key Length", func(t *testing.T) {
		ciphertext, nonce, err := encryptor.Encrypt(testEncryptableData, testStorageKey)
		assert.NoError(t, err, "Encryption should succeed")

		shortKey := make([]byte, 16)
		_, err = encryptor.Decrypt(ciphertext, shortKey, nonce)
		assert.Error(t, err, "Decryption with short key should fail")
		assert.Contains(t, err.Error(), "key must be 32 bytes long")
	})

	t.Run("Decryption with Incorrect Nonce", func(t *testing.T) {
		ciphertext, _, err := encryptor.Encrypt(testEncryptableData, testStorageKey)
		assert.NoError(t, err, "Encryption should succeed")

		incorrectNonce := make([]byte, 16)
		_, err = encryptor.Decrypt(ciphertext, testStorageKey, incorrectNonce)
		assert.Error(t, err, "Decryption with incorrect nonce should fail")
		assert.Contains(t, err.Error(), "incorrect nonce size")
	})
}
