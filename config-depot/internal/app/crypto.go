package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

const uploadIVSize = aes.BlockSize

func EncryptUploadBundle(plaintext []byte, key []byte) ([]byte, error) {
	iv := make([]byte, uploadIVSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate upload IV: %w", err)
	}
	ciphertext, err := EncryptAES256CBC(plaintext, key, iv)
	if err != nil {
		return nil, err
	}

	encrypted := make([]byte, 0, len(iv)+len(ciphertext))
	encrypted = append(encrypted, iv...)
	encrypted = append(encrypted, ciphertext...)
	return encrypted, nil
}

func DecryptUploadBundle(encrypted []byte, key []byte) ([]byte, error) {
	if len(encrypted) <= uploadIVSize {
		return nil, errors.New("encrypted upload bundle is too short")
	}
	iv := encrypted[:uploadIVSize]
	ciphertext := encrypted[uploadIVSize:]
	return DecryptAES256CBC(ciphertext, key, iv)
}

func DecryptAES256CBC(ciphertext []byte, key []byte, iv []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("AES-256 key must be 32 bytes, got %d", len(key))
	}
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("AES-CBC IV must be %d bytes, got %d", aes.BlockSize, len(iv))
	}
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext length must be a non-zero multiple of AES block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)

	return unpadPKCS7(plaintext, aes.BlockSize)
}

func EncryptAES256CBC(plaintext []byte, key []byte, iv []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("AES-256 key must be 32 bytes, got %d", len(key))
	}
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("AES-CBC IV must be %d bytes, got %d", aes.BlockSize, len(iv))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	padded := padPKCS7(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)
	return ciphertext, nil
}

func DecodeBase64Secret(value string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode base64 secret: %w", err)
	}
	if len(decoded) != 32 {
		return nil, fmt.Errorf("base64 secret must decode to 32 bytes, got %d", len(decoded))
	}
	return decoded, nil
}

func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}
	return padded
}

func unpadPKCS7(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid PKCS#7 data length")
	}

	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, errors.New("invalid PKCS#7 padding")
	}
	for _, value := range data[len(data)-padding:] {
		if int(value) != padding {
			return nil, errors.New("invalid PKCS#7 padding bytes")
		}
	}
	return data[:len(data)-padding], nil
}
