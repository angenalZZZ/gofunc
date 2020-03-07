package f

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// CryptoAesCBCEncrypt aes CBC模式+key(16bytes)+iv(16bytes).
// encryptedString := hex.EncodeToString(encryptedBytes)
// encryptedString := base64.StdEncoding.EncodeToString(encryptedBytes)
func CryptoAesCBCEncrypt(origData, key, iv []byte) ([]byte, error) {
	if key == nil || len(key)%16 != 0 {
		return nil, errors.New("wrong key")
	}
	if iv == nil || len(iv)%16 != 0 {
		return nil, errors.New("wrong iv")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv[:blockSize])
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)
	return encrypted, nil
}

// CryptoAesCBCEncryptWithHmacSHA1 aes CBC模式+key(16bytes)+salt(8bytes)+iv(16bytes), iterations 1000, output 32 bytes.
func CryptoAesCBCEncryptWithHmacSHA1(origData, key, salt, iv []byte, iterations, outLen int) ([]byte, error) {
	password := CryptoSecretKeyPBKDF2WithHmacSHA1(key, salt, iterations, outLen)
	return CryptoAesCBCEncrypt(origData, password[0:32], iv)
}

// CryptoAesCBCEncryptWithHmacSHA256 aes CBC模式+key(16bytes)+salt(8bytes)+iv(16bytes), iterations 1000, output 32 bytes.
func CryptoAesCBCEncryptWithHmacSHA256(origData, key, salt, iv []byte, iterations, outLen int) ([]byte, error) {
	password := CryptoSecretKeyPBKDF2WithHmacSHA256(key, salt, iterations, outLen)
	return CryptoAesCBCEncrypt(origData, password[0:32], iv)
}

// CryptoAesCBCDecrypt aes CBC模式+key(16bytes)+iv(16bytes).
// encryptedBytes, err := hex.DecodeString(encryptedString)
// encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedString)
func CryptoAesCBCDecrypt(encrypted, key, iv []byte) ([]byte, error) {
	if key == nil || len(key)%16 != 0 {
		return nil, errors.New("wrong key")
	}
	if iv == nil || len(iv)%16 != 0 {
		return nil, errors.New("wrong iv")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(encrypted))
	blockMode.CryptBlocks(origData, encrypted)
	origData = pkcs5UnPadding(origData)
	return origData, nil
}

// CryptoAesCBCDecryptWithHmacSHA1 aes CBC模式+key(16bytes)+salt(8bytes)+iv(16bytes), iterations 1000, output 32 bytes.
func CryptoAesCBCDecryptWithHmacSHA1(encrypted, key, salt, iv []byte, iterations, outLen int) ([]byte, error) {
	password := CryptoSecretKeyPBKDF2WithHmacSHA1(key, salt, iterations, outLen)
	return CryptoAesCBCDecrypt(encrypted, password[0:32], iv)
}

// CryptoAesCBCDecryptWithHmacSHA256 aes CBC模式+key(16bytes)+salt(8bytes)+iv(16bytes), iterations 1000, output 32 bytes.
func CryptoAesCBCDecryptWithHmacSHA256(encrypted, key, salt, iv []byte, iterations, outLen int) ([]byte, error) {
	password := CryptoSecretKeyPBKDF2WithHmacSHA256(key, salt, iterations, outLen)
	return CryptoAesCBCDecrypt(encrypted, password[0:32], iv)
}
