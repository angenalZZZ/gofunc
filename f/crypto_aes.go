package f

import (
	"crypto/aes"
	"crypto/cipher"
)

// CryptoAesCBCEncrypt AES/CBC/PKCS5Padding+key(16/24/32bytes)+iv(16/24/32bytes)-Encrypt.
// encryptedBytes := CryptoAesCBCEncrypt(origData, key, iv)
// encryptedString := hex.EncodeToString(encryptedBytes)
// encryptedString := base64.StdEncoding.EncodeToString(encryptedBytes)
func CryptoAesCBCEncrypt(origData, key, iv []byte) []byte {
	MustBytes(key, 16, 24, 32)
	MustBytes(iv, 16, 24, 32)
	blockMode, blockSize := CryptoAesBlockMode(key, iv, cipher.NewCBCEncrypter)
	origData = CryptoPKCS5Padding(origData, blockSize)
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)
	return encrypted
}

// CryptoAesCBCEncryptWithHmacSHA1 AES/CBC/PKCS5Padding+key(16/24/32bytes)+salt(8bytes)+iv(16/24/32bytes)+iterations(1000)+outLen(32).
func CryptoAesCBCEncryptWithHmacSHA1(origData, key, salt, iv []byte, iterations, outLen int) []byte {
	password := CryptoSecretKeyPBKDF2WithHmacSHA1(key, salt, iterations, outLen)
	return CryptoAesCBCEncrypt(origData, password[0:32], iv)
}

// CryptoAesCBCEncryptWithHmacSHA256 AES/CBC/PKCS5Padding+key(16/24/32bytes)+salt(8bytes)+iv(16/24/32bytes)+iterations(1000)+outLen(32).
func CryptoAesCBCEncryptWithHmacSHA256(origData, key, salt, iv []byte, iterations, outLen int) []byte {
	password := CryptoSecretKeyPBKDF2WithHmacSHA256(key, salt, iterations, outLen)
	return CryptoAesCBCEncrypt(origData, password[0:32], iv)
}

// CryptoAesCBCDecrypt aes AES/CBC/PKCS5Padding+key(16/24/32bytes)+iv(16/24/32bytes)-Decrypt.
// encryptedBytes, err := hex.DecodeString(encryptedString)
// encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedString)
// origData := CryptoAesCBCEncrypt(encryptedBytes, key, iv)
func CryptoAesCBCDecrypt(encrypted, key, iv []byte) []byte {
	MustBytes(key, 16, 24, 32)
	MustBytes(iv, 16, 24, 32)
	blockMode, _ := CryptoAesBlockMode(key, iv, cipher.NewCBCDecrypter)
	origData := make([]byte, len(encrypted))
	blockMode.CryptBlocks(origData, encrypted)
	return CryptoPKCS5UnPadding(origData)
}

// CryptoAesCBCDecryptWithHmacSHA1 AES/CBC/PKCS5Padding+key(16/24/32bytes)+salt(8bytes)+iv(16/24/32bytes)+iterations(1000)+outLen(32).
func CryptoAesCBCDecryptWithHmacSHA1(encrypted, key, salt, iv []byte, iterations, outLen int) []byte {
	password := CryptoSecretKeyPBKDF2WithHmacSHA1(key, salt, iterations, outLen)
	return CryptoAesCBCDecrypt(encrypted, password[0:32], iv)
}

// CryptoAesCBCDecryptWithHmacSHA256 AES/CBC/PKCS5Padding+key(16/24/32bytes)+salt(8bytes)+iv(16/24/32bytes)+iterations(1000)+outLen(32).
func CryptoAesCBCDecryptWithHmacSHA256(encrypted, key, salt, iv []byte, iterations, outLen int) []byte {
	password := CryptoSecretKeyPBKDF2WithHmacSHA256(key, salt, iterations, outLen)
	return CryptoAesCBCDecrypt(encrypted, password[0:32], iv)
}

// CryptoAesBlockMode New Aes BlockMode Method+key(16/24/32bytes)+iv(16/24/32bytes).
// blockMode, blockSize := CryptoAesBlockMode(key, iv, cipher.NewCBCEncrypter)
// blockMode, blockSize := CryptoAesBlockMode(key, iv, cipher.NewCBCDecrypter)
func CryptoAesBlockMode(key, iv []byte, f func(cipher.Block, []byte) cipher.BlockMode) (cipher.BlockMode, int) {
	block, err := aes.NewCipher(key)
	Must(err)
	blockSize := block.BlockSize()
	return f(block, iv[:blockSize]), blockSize
}

// CryptoAesStream New Aes Stream Method+key(16/24/32bytes)+iv(16/24/32bytes).
// stream := CryptoAesStream(key, iv, cipher.NewCFBDecrypter)
// stream := CryptoAesStream(key, iv, cipher.NewCFBEncrypter)
func CryptoAesStream(key, iv []byte, f func(cipher.Block, []byte) cipher.Stream) cipher.Stream {
	block, err := aes.NewCipher(key)
	Must(err)
	return f(block, iv)
}
