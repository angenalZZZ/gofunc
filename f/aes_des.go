package f

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"hash"
)

// EncryptAes aes CBC加密+key
func EncryptAes(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)
	return encrypted, nil
}

// EncryptAes128 aes CBC加密+key(16字节)+iv(16字节)
func EncryptAes128(origData, key, iv []byte) ([]byte, error) {
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

// EncryptAes128s aes CBC加密+key(16字节)+salt(8字节)+iv(16字节)
func EncryptAes128s(origData, key, salt, iv []byte) ([]byte, error) {
	password := Pbkdf2Rfc2898DeriveBytes(key, salt)
	return EncryptAes128(origData, password[0:32], iv)
}

// DecryptAes aes CBC解密+key
// encrypted, err := hex.DecodeString(encryptedString)
func DecryptAes(encrypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(encrypted))
	blockMode.CryptBlocks(origData, encrypted)
	origData = pkcs5UnPadding(origData)
	return origData, nil
}

// DecryptAes128 aes CBC解密+key(16字节)+iv(16字节)
// encrypted, err := hex.DecodeString(encryptedString)
func DecryptAes128(encrypted, key, iv []byte) ([]byte, error) {
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

// DecryptAes128s aes CBC解密+key(16字节)+salt(8字节)+iv(16字节)
func DecryptAes128s(encrypted, key, salt, iv []byte) ([]byte, error) {
	password := Pbkdf2Rfc2898DeriveBytes(key, salt)
	return DecryptAes128(encrypted, password[0:32], iv)
}

// EncryptDes128 des CBC加密+key(8字节)+iv(16字节)
func EncryptDes128(origData, key, iv []byte) ([]byte, error) {
	if key == nil || len(key)%8 != 0 {
		return nil, errors.New("wrong key")
	}
	if iv == nil || len(iv)%8 != 0 {
		return nil, errors.New("wrong iv")
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv[:blockSize])
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)
	return encrypted, nil
	//hex.EncodeToString(encrypted)
}

// DecryptDes128 des CBC解密+key(8字节)+iv(16字节)
// encrypted, err := hex.DecodeString(encryptedString)
func DecryptDes128(encrypted, key, iv []byte) ([]byte, error) {
	if key == nil || len(key)%8 != 0 {
		return nil, errors.New("wrong key")
	}
	if iv == nil || len(iv)%8 != 0 {
		return nil, errors.New("wrong iv")
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	mode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(encrypted))
	mode.CryptBlocks(origData, encrypted)
	origData = pkcs5UnPadding(origData)
	return origData, nil
}

// EncryptDesECB des ECB加密
func EncryptDesECB(origData, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs5Padding(origData, blockSize)
	if len(origData)%blockSize != 0 {
		return nil, errors.New("need a multiple of the block size")
	}
	out := make([]byte, len(origData))
	dst := out
	for len(origData) > 0 {
		block.Encrypt(dst, origData[:blockSize])
		origData = origData[blockSize:]
		dst = dst[blockSize:]
	}
	return out, nil
}

// DecryptDesECB des ECB解密
// encrypted, err := hex.DecodeString(encryptedString)
func DecryptDesECB(encrypted, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(encrypted)%blockSize != 0 {
		return nil, errors.New("need a multiple of the block size")
	}
	out := make([]byte, len(encrypted))
	dst := out
	for len(encrypted) > 0 {
		block.Decrypt(dst, encrypted[:blockSize])
		encrypted = encrypted[blockSize:]
		dst = dst[blockSize:]
	}
	out = pkcs5UnPadding(out)
	return out, nil
}

// EncryptTripleDesECB des ECB加密+key(24字节) 三重加密
func EncryptTripleDesECB(origData, key []byte) ([]byte, error) {
	if key == nil || len(key) != 24 {
		return nil, errors.New("wrong key")
	}
	tk := make([]byte, 24, 24)
	copy(tk, key)
	k1 := tk[:8]
	k2 := tk[8:16]
	k3 := tk[16:]
	block, err := des.NewCipher(k1)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs5Padding(origData, blockSize)
	buf1, err1 := desEncrypt(origData, k1)
	if err1 != nil {
		return nil, err1
	}
	buf2, err2 := desDecrypt(buf1, k2)
	if err2 != nil {
		return nil, err2
	}
	buf3, err3 := desEncrypt(buf2, k3)
	if err3 != nil {
		return nil, err3
	}
	return buf3, nil
}

func desEncrypt(origData, key []byte) ([]byte, error) {
	if len(origData) < 1 || len(key) < 1 {
		return nil, errors.New("wrong data or key")
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	if len(origData)%bs != 0 {
		return nil, errors.New("wrong padding")
	}
	out := make([]byte, len(origData))
	dst := out
	for len(origData) > 0 {
		block.Encrypt(dst, origData[:bs])
		origData = origData[bs:]
		dst = dst[bs:]
	}
	return out, nil
}

func desDecrypt(encrypted, key []byte) ([]byte, error) {
	if len(encrypted) < 1 || len(key) < 1 {
		return nil, errors.New("wrong data or key")
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(encrypted))
	dst := out
	bs := block.BlockSize()
	if len(encrypted)%bs != 0 {
		return nil, errors.New("wrong encrypted size")
	}
	for len(encrypted) > 0 {
		block.Decrypt(dst, encrypted[:bs])
		encrypted = encrypted[bs:]
		dst = dst[bs:]
	}
	return out, nil
}

func pkcs5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

func pkcs7UnPadding(origData []byte, blockSize int) []byte {
	return origData[:len(origData)-int(origData[len(origData)-1])]
}

// Pbkdf2Rfc2898DeriveBytes a key provided password, salt.
func Pbkdf2Rfc2898DeriveBytes(password, salt []byte) []byte {
	return Pbkdf2WithHMAC(sha256.New, password, salt, 9999, 64)
}

// Pbkdf2WithHMAC derives key of length outLen from the provided password, salt,
// and the number of iterations using PKCS#5 PBKDF2 with the provided
// hash function in HMAC.
//
// Caller is responsible to make sure that outLen < (2^32-1) * hash.Size().
func Pbkdf2WithHMAC(hash func() hash.Hash, password []byte, salt []byte, iterations int, outLen int) []byte {
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
