package f

import (
	"crypto/cipher"
	"crypto/des"
)

// CryptoDesCBCEncrypt DES/CBC/PKCS5Padding+key(8bytes)+iv(16bytes)-Encrypt.
// encryptedBytes := CryptoDesCBCEncrypt(origData, key, iv)
// encryptedString := hex.EncodeToString(encryptedBytes)
// encryptedString := base64.StdEncoding.EncodeToString(encryptedBytes)
func CryptoDesCBCEncrypt(origData, key, iv []byte) []byte {
	MustBytes(key, 8)
	MustBytes(iv, 16)
	block, err := des.NewCipher(key)
	Must(err)
	bs := block.BlockSize()
	origData = CryptoPKCS5Padding(origData, bs)
	blockMode := cipher.NewCBCEncrypter(block, iv[:bs])
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)
	return encrypted
}

// CryptoDesCBCDecrypt DES/CBC/PKCS5Padding+key(8bytes)+iv(16bytes)-Decrypt.
// encryptedBytes, err := hex.DecodeString(encryptedString)
// encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedString)
// origData := CryptoDesCBCDecrypt(encryptedBytes, key, iv)
func CryptoDesCBCDecrypt(encrypted, key, iv []byte) []byte {
	MustBytes(key, 8)
	MustBytes(iv, 16)
	block, err := des.NewCipher(key)
	Must(err)
	bs := block.BlockSize()
	mode := cipher.NewCBCDecrypter(block, iv[:bs])
	origData := make([]byte, len(encrypted))
	mode.CryptBlocks(origData, encrypted)
	return CryptoPKCS5UnPadding(origData)
}

// CryptoDesECBEncrypt DES/ECB/PKCS5Padding+key(8bytes)-Encrypt.
func CryptoDesECBEncrypt(origData, key []byte) []byte {
	MustBytes(key, 8)
	block, err := des.NewCipher(key)
	Must(err)
	bs := block.BlockSize()
	origData = CryptoPKCS5Padding(origData, bs)
	if len(origData)%bs != 0 {
		return nil
	}
	out := make([]byte, len(origData))
	dst := out
	for len(origData) > 0 {
		block.Encrypt(dst, origData[:bs])
		origData = origData[bs:]
		dst = dst[bs:]
	}
	return out
}

// CryptoDesECBDecrypt DES/ECB/PKCS5Padding+key(8bytes)-Decrypt.
func CryptoDesECBDecrypt(encrypted, key []byte) []byte {
	MustBytes(key, 8)
	block, err := des.NewCipher(key)
	Must(err)
	bs := block.BlockSize()
	if len(encrypted)%bs != 0 {
		return nil
	}
	out := make([]byte, len(encrypted))
	dst := out
	for len(encrypted) > 0 {
		block.Decrypt(dst, encrypted[:bs])
		encrypted = encrypted[bs:]
		dst = dst[bs:]
	}
	return CryptoPKCS5UnPadding(out)
}

// CryptoDesECBTripleEncrypt DES/ECB/PKCS5Padding+key(24bytes)+Tripled-Encrypt.
func CryptoDesECBTripleEncrypt(origData, key []byte) []byte {
	MustBytes(key, 24)
	tk := make([]byte, 24, 24)
	copy(tk, key)
	k1 := tk[:8]
	k2 := tk[8:16]
	k3 := tk[16:]
	block, err := des.NewCipher(k1)
	Must(err)
	bs := block.BlockSize()
	origData = CryptoPKCS5Padding(origData, bs)
	buf1 := cryptoDesECBEncrypt(origData, k1)
	buf2 := cryptoDesECBDecrypt(buf1, k2)
	buf3 := cryptoDesECBEncrypt(buf2, k3)
	return buf3
}

func cryptoDesECBEncrypt(origData, key []byte) []byte {
	block, err := des.NewCipher(key)
	Must(err)
	bs := block.BlockSize()
	if len(origData)%bs != 0 {
		return nil
	}
	out := make([]byte, len(origData))
	dst := out
	for len(origData) > 0 {
		block.Encrypt(dst, origData[:bs])
		origData = origData[bs:]
		dst = dst[bs:]
	}
	return out
}

func cryptoDesECBDecrypt(encrypted, key []byte) []byte {
	block, err := des.NewCipher(key)
	Must(err)
	out := make([]byte, len(encrypted))
	dst := out
	bs := block.BlockSize()
	if len(encrypted)%bs != 0 {
		return nil
	}
	for len(encrypted) > 0 {
		block.Decrypt(dst, encrypted[:bs])
		encrypted = encrypted[bs:]
		dst = dst[bs:]
	}
	return out
}
