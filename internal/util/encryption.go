package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

func GenerateRandomAESKey() (string, []byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", nil, err
	}
	return base64.StdEncoding.EncodeToString(key), key, err
}

func EncryptAESGCM(plainText string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nil, nonce, []byte(plainText), nil)

	encrypted := append(nonce, cipherText...)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func DecryptAESGCM(encryptedText string, key []byte) (string, error) {
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(encryptedBytes) < nonceSize {
		return "", errors.New("encrypted str to short")
	}

	nonce, cipherText := encryptedBytes[:nonceSize], encryptedBytes[nonceSize:]

	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func Base64Decode(str string) []byte {
	decode, _ := base64.StdEncoding.DecodeString(str)
	return decode
}

func EncryptAESGCMStr(plainText string, key []byte) string {
	enc, _ := EncryptAESGCM(plainText, key)
	return enc
}

func DecryptAESGCMStr(encryptedText string, key []byte) string {
	dec, _ := DecryptAESGCM(encryptedText, key)
	return dec
}
