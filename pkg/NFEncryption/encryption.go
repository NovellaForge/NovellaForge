package NFEncryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// Encrypt takes a plaintext byte array and encrypts it using the provided key.
// It returns the ciphertext and an error if one occurs.
// It generates the cipher text using AES-GCM and a nonce.
// A Nonce is a number that can only be used once and will be unique for each encryption, but it is not a secret
// It is safe to store the nonce alongside the ciphertext
// The nonce is prepended to the ciphertext
func Encrypt(plaintext []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher(validateKey([]byte(key), 32))
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	newGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := newGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt takes a ciphertext byte array and decrypts it using the provided key
func Decrypt(ciphertext []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher(validateKey([]byte(key), 32))
	if err != nil {
		return nil, err
	}

	newGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := newGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := newGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// validateKey takes a key and a length and returns a key of the correct length by truncating or padding it
func validateKey(key []byte, length int) []byte {
	if len(key) == length {
		return key
	}
	newKey := make([]byte, length)
	copy(newKey, key)
	return newKey
}
