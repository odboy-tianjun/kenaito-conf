package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"kenaito-conf/server/internal/config"
	"strings"
)

const CryptPrefix = "{Cipher}"

type CryptoService struct{}

func NewCryptoService() *CryptoService {
	return &CryptoService{}
}

func (s *CryptoService) GetCryptKey(appId int, env string) string {
	base := config.Cfg.Crypto.SecretKey + fmt.Sprintf("%x", appId) + env
	h := md5.Sum([]byte(base))
	return hex.EncodeToString(h[:])
}

func (s *CryptoService) Encrypt(appId int, env string, data string) string {
	key := s.GetCryptKey(appId, env)
	data = strings.ReplaceAll(data, CryptPrefix, "")
	encrypted := encryptWithAES(key, data)
	return CryptPrefix + encrypted
}

func (s *CryptoService) IsEncrypted(appId int, env string, data string) bool {
	if !strings.HasPrefix(data, CryptPrefix) {
		return false
	}
	key := s.GetCryptKey(appId, env)
	_, err := decryptWithAES(key, strings.TrimPrefix(data, CryptPrefix))
	return err == nil
}

func (s *CryptoService) Decrypt(secret string, data string) (string, error) {
	return decryptWithAES(secret, data)
}

func encryptWithAES(key string, data string) string {
	secretKey := md5.Sum([]byte(key))
	aesKey := hex.EncodeToString(secretKey[:])[16:]

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return ""
	}

	// ECB mode (compatible with Java version)
	plainText := pkcs7Pad([]byte(data), aes.BlockSize)
	encrypted := make([]byte, len(plainText))
	mode := cipher.NewCBCEncrypter(block, make([]byte, aes.BlockSize))
	mode.CryptBlocks(encrypted, plainText)

	return base64.StdEncoding.EncodeToString(encrypted)
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

	// ECB mode (compatible with Java version)
	plainText := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, make([]byte, aes.BlockSize))
	mode.CryptBlocks(plainText, cipherText)

	plainText, err = pkcs7Unpad(plainText, aes.BlockSize)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
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
