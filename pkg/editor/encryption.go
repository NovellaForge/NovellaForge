package editor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

const MainEncryptionKey = "NovellaForge+"

func encrypt(plainText, key []byte) (string, error) {
	//combine the main encryption key with the key provided
	key = append([]byte(MainEncryptionKey), key...)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plainTextBytes := plainText
	cipherText := make([]byte, aes.BlockSize+len(plainTextBytes))

	// Initialization Vector
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// Encrypt using CBC mode
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainTextBytes)

	// Return as base64 string
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func decrypt(cipherTextBase64 string, key []byte) (string, error) {
	//combine the main encryption key with the key provided
	key = append([]byte(MainEncryptionKey), key...)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	cipherText, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return "", err
	}

	// Initialization Vector
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// Decrypt using CBC mode
	plainText := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plainText, cipherText)

	return string(plainText), nil
}
