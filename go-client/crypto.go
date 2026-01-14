package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

const CryptPrefix = "{Cipher}"

func decryptConfig(secret string, data string) (string, error) {
	if !strings.HasPrefix(data, CryptPrefix) {
		return data, nil
	}
	encrypted := strings.TrimPrefix(data, CryptPrefix)
	return decryptWithAES(secret, encrypted)
}

func decryptWithAES(key string, data string) (string, error) {
	secretKey := md5.Sum([]byte(key))
	aesKey := hex.EncodeToString(secretKey[:])[16:]

	cipherText, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return "", err
	}

	if len(cipherText)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of block size")
	}

	plainText := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, make([]byte, aes.BlockSize))
	mode.CryptBlocks(plainText, cipherText)

	plainText, err = pkcs7Unpad(plainText, aes.BlockSize)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > blockSize || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	return data[:len(data)-padding], nil
}
