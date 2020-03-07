package f

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"hash"
)

// CryptoSecretKeyPBKDF2WithHmacSHA1 derives key provided password, salt, iterations 1000, output 32 bytes.
func CryptoSecretKeyPBKDF2WithHmacSHA1(password, salt []byte, iterations, outLen int) []byte {
	return CryptoSecretKeyPBKDF2WithHmac(sha1.New, password, salt, iterations, outLen)
}

// CryptoSecretKeyPBKDF2WithHmacSHA256 derives key provided password, salt, iterations 1000, output 32 bytes.
func CryptoSecretKeyPBKDF2WithHmacSHA256(password, salt []byte, iterations, outLen int) []byte {
	return CryptoSecretKeyPBKDF2WithHmac(sha256.New, password, salt, iterations, outLen)
}

// CryptoSecretKeyPBKDF2WithHmac derives key of length outLen from the provided password, salt,
// and the number of iterations using PKCS#5 PBKDF2 with the provided hash function in HMAC.
//
// Caller is responsible to make sure that outLen < (2^32-1) * hash.Size().
func CryptoSecretKeyPBKDF2WithHmac(hash func() hash.Hash, password, salt []byte, iterations, outLen int) []byte {
	out := make([]byte, outLen)
	hashSize := hash().Size()
	buf := make([]byte, 4)
	block := 1
	p := out
	for outLen > 0 {
		clean := outLen
		if clean > hashSize {
			clean = hashSize
		}
		buf[0] = byte((block >> 24) & 0xff)
		buf[1] = byte((block >> 16) & 0xff)
		buf[2] = byte((block >> 8) & 0xff)
		buf[3] = byte((block) & 0xff)
		hmacPass := hmac.New(hash, password)
		hmacPass.Write(salt)
		hmacPass.Write(buf)
		tmp := hmacPass.Sum(nil)
		for i := 0; i < clean; i++ {
			p[i] = tmp[i]
		}
		for j := 1; j < iterations; j++ {
			hmacPass.Reset()
			hmacPass.Write(tmp)
			tmp = hmacPass.Sum(nil)
			for k := 0; k < clean; k++ {
				p[k] ^= tmp[k]
			}
		}
		outLen -= clean
		block++
		p = p[clean:]
	}
	return out
}

// pkcs5Padding for AES/CBC/PKCS5Padding
func pkcs5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

// pkcs5UnPadding for AES/CBC/PKCS5UnPadding
func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

//func pkcs7UnPadding(origData []byte, blockSize int) []byte {
//	return origData[:len(origData)-int(origData[len(origData)-1])]
//}
