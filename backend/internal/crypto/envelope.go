package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// EncryptedSecret holds everything needed to decrypt a secret value.
// All fields are stored in the database; the plaintext and DEK are never stored.
type EncryptedSecret struct {
	EncryptedValue []byte // AES-256-GCM ciphertext of the secret value
	DataNonce      []byte // 12-byte nonce for the value encryption
	EncryptedDEK   []byte // AES-256-GCM ciphertext of the DEK
	DEKNonce       []byte // 12-byte nonce for the DEK encryption
}

// Encrypt encrypts a plaintext secret value using envelope encryption.
// kek must be exactly 32 bytes (AES-256).
func Encrypt(kek []byte, plaintext string) (*EncryptedSecret, error) {
	if len(kek) != 32 {
		return nil, errors.New("crypto: KEK must be exactly 32 bytes")
	}

	// Step 1: Generate a fresh random DEK for this secret.
	// Each secret gets its own DEK.
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, err
	}

	// Step 2: Encrypt the plaintext value with the DEK.
	encryptedValue, dataNonce, err := aesgcmEncrypt(dek, []byte(plaintext))
	if err != nil {
		return nil, err
	}

	// Step 3: Encrypt the DEK with the master KEK.
	// Now the KEK only needs to touch the DEK, never the raw secret.
	encryptedDEK, dekNonce, err := aesgcmEncrypt(kek, dek)
	if err != nil {
		return nil, err
	}

	return &EncryptedSecret{
		EncryptedValue: encryptedValue,
		DataNonce:      dataNonce,
		EncryptedDEK:   encryptedDEK,
		DEKNonce:       dekNonce,
	}, nil
}

// Decrypt reverses the envelope encryption to recover the plaintext secret.
func Decrypt(kek []byte, es *EncryptedSecret) (string, error) {
	if len(kek) != 32 {
		return "", errors.New("crypto: KEK must be exactly 32 bytes")
	}

	// Step 1: Recover the DEK by decrypting it with the master KEK.
	dek, err := aesgcmDecrypt(kek, es.EncryptedDEK, es.DEKNonce)
	if err != nil {
		return "", errors.New("crypto: failed to decrypt DEK (wrong KEK?)")
	}

	// Step 2: Use the recovered DEK to decrypt the actual secret value.
	plaintext, err := aesgcmDecrypt(dek, es.EncryptedValue, es.DataNonce)
	if err != nil {
		return "", errors.New("crypto: failed to decrypt secret value")
	}

	return string(plaintext), nil
}

// aesgcmEncrypt encrypts plaintext with the given 32-byte key using AES-256-GCM.
// Returns the ciphertext and the randomly generated nonce separately.
func aesgcmEncrypt(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	// GCM requires a unique nonce for every encryption with the same key.
	// Using a random nonce (rather than a counter) makes it safe to call
	// this function many times without coordination between callers.
	nonce = make([]byte, gcm.NonceSize()) // 12 bytes for AES-GCM
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	// Seal encrypts and appends the authentication tag.
	// The tag detects tampering: if anyone flips a bit in ciphertext, Seal will fail.
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// aesgcmDecrypt decrypts ciphertext produced by aesgcmEncrypt.
func aesgcmDecrypt(key, ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Open decrypts and verifies the authentication tag in one step.
	// If either the ciphertext or the tag was modified, this returns an error.
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
