package f

import (
	"crypto/cipher"
	"crypto/des"
	"errors"
)

// CryptoDesCBCEncrypt des CBC模式+key(8bytes)+iv(16bytes).
// encryptedString := hex.EncodeToString(encryptedBytes)
// encryptedString := base64.StdEncoding.EncodeToString(encryptedBytes)
func CryptoDesCBCEncrypt(origData, key, iv []byte) ([]byte, error) {
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
	origData = CryptoPKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv[:blockSize])
	encrypted := make([]byte, len(origData))
	blockMode.CryptBlocks(encrypted, origData)
	return encrypted, nil
}

// CryptoDesCBCDecrypt des CBC模式+key(8bytes)+iv(16bytes).
// encryptedBytes, err := hex.DecodeString(encryptedString)
// encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedString)
func CryptoDesCBCDecrypt(encrypted, key, iv []byte) ([]byte, error) {
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
	origData = CryptoPKCS5UnPadding(origData)
	return origData, nil
}

// CryptoDesECBEncrypt des ECB模式+key(8bytes).
func CryptoDesECBEncrypt(origData, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = CryptoPKCS5Padding(origData, blockSize)
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

// CryptoDesECBDecrypt des ECB模式+key(8bytes).
func CryptoDesECBDecrypt(encrypted, key []byte) ([]byte, error) {
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
	out = CryptoPKCS5UnPadding(out)
	return out, nil
}

// CryptoDesECBTripleEncrypt des ECB模式+key(24bytes)+Triple三重加密.
func CryptoDesECBTripleEncrypt(origData, key []byte) ([]byte, error) {
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
	origData = CryptoPKCS5Padding(origData, blockSize)
	buf1, err1 := cryptoDesECBEncrypt(origData, k1)
	if err1 != nil {
		return nil, err1
	}
	buf2, err2 := cryptoDesECBDecrypt(buf1, k2)
	if err2 != nil {
		return nil, err2
	}
	buf3, err3 := cryptoDesECBEncrypt(buf2, k3)
	if err3 != nil {
		return nil, err3
	}
	return buf3, nil
}

func cryptoDesECBEncrypt(origData, key []byte) ([]byte, error) {
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

func cryptoDesECBDecrypt(encrypted, key []byte) ([]byte, error) {
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
