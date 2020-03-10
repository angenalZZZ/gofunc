package f

import (
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"testing"
)

func TestMD5(t *testing.T) {
	origData, encryptedString := "hello", "5d41402abc4b2a76b9719d911017c592"
	encryptedStringGo := CryptoMD5(origData)
	if encryptedString != encryptedStringGo {
		t.Log(origData)
		t.Log(encryptedStringGo)
		t.Fatal(" encryptedString != encryptedStringGo ")
	}
}

func TestCryptoAes(t *testing.T) {
	origData := []byte("hello")
	key := []byte("TmIhgugCGFpU7S3v")
	iv := []byte("jkE49230Tf093b42")
	encryptedBytes := CryptoAesCBCEncrypt(origData, key, iv)
	encryptedString := hex.EncodeToString(encryptedBytes)
	// Output: CryptoAesCBCEncrypt: hello => 548e8841b4baa92451bc4e7fd875ad1c
	t.Logf("CryptoAesCBCEncrypt: %s => %s", origData, encryptedString)
	encryptedRaw, err3 := hex.DecodeString(encryptedString)
	if err3 != nil {
		t.Fatal(err3)
	}
	if string(encryptedRaw) != string(encryptedBytes) {
		t.Fatal("encryptedRaw != encryptedBytes")
	}
	origDataRaw := CryptoAesCBCDecrypt(encryptedRaw, key, iv)
	if string(origDataRaw) != string(origData) {
		t.Fatal("origDataRaw != origData")
	}
	// Output: CryptoAesCBCDecrypt: 548e8841b4baa92451bc4e7fd875ad1c => hello
	t.Logf("CryptoAesCBCDecrypt: %s => %s", hex.EncodeToString(encryptedRaw), origDataRaw)

	// Crypto: pbkdf2 Rfc2898DeriveBytes password
	salt := []byte("hgt!16kl")
	encrypted1 := CryptoAesCBCEncryptWithHmacSHA1(origData, key, salt, iv, 1000, 32)
	encryptedString1 := hex.EncodeToString(encrypted1)
	// Output: CryptoAesCBCEncryptWithHmacSHA1: hello => 7a940d1245ad99fa6cbb4c6fe72f2ed8
	t.Logf("CryptoAesCBCEncryptWithHmacSHA1: %s => %s", origData, encryptedString1)
	encryptedRaw1, err31 := hex.DecodeString(encryptedString1)
	if err31 != nil {
		t.Fatal(err31)
	}
	if string(encryptedRaw1) != string(encrypted1) {
		t.Fatal("encryptedRaw1 != encrypted1")
	}
	origDataRaw1 := CryptoAesCBCDecryptWithHmacSHA1(encryptedRaw1, key, salt, iv, 1000, 32)
	if string(origDataRaw1) != string(origData) {
		t.Fatal("origDataRaw1 != origData")
	}
	// Output: CryptoAesCBCDecryptWithHmacSHA1: 7a940d1245ad99fa6cbb4c6fe72f2ed8 => hello
	t.Logf("CryptoAesCBCDecryptWithHmacSHA1: %s => %s", hex.EncodeToString(encryptedRaw1), origDataRaw1)
}

func TestCryptoDes(t *testing.T) {
	origData := []byte("hello")
	key := []byte("GFpU7S3v")
	iv := []byte("jkE49230Tf093b42")
	encryptedBytes := CryptoDesCBCEncrypt(origData, key, iv)
	encryptedString := hex.EncodeToString(encryptedBytes)
	// Output: CryptoDesCBCEncrypt: hello => 898aff98549d75cb
	t.Logf("CryptoDesCBCEncrypt: %s => %s", origData, encryptedString)
	encryptedRaw, err3 := hex.DecodeString(encryptedString)
	if err3 != nil {
		t.Fatal(err3)
	}
	if string(encryptedRaw) != string(encryptedBytes) {
		t.Fatal("encryptedRaw != encryptedBytes")
	}
	origDataRaw := CryptoDesCBCDecrypt(encryptedRaw, key, iv)
	if string(origDataRaw) != string(origData) {
		t.Fatal("origDataRaw != origData")
	}
	// Output: CryptoDesCBCDecrypt: 898aff98549d75cb => hello
	t.Logf("CryptoDesCBCDecrypt: %s => %s", hex.EncodeToString(encryptedRaw), origDataRaw)

	// Crypto: des ECB Triple Encrypt
	key = []byte("TmIhgugCGFpU7S3vGFpU7S3v")
	encrypted := CryptoDesECBTripleEncrypt(origData, key)
	encryptedString = hex.EncodeToString(encrypted)
	// Output: CryptoDesECBTripleEncrypt: hello => 86f21066c5ba8c49
	t.Logf("CryptoDesECBTripleEncrypt: %s => %s", origData, encryptedString)
}

func TestCryptoRSA(t *testing.T) {
	origData := `{"Customer":"gbxy","SecretIdCard":"9c1c0dd59ff33f9ac37bd072ac2df86d","Timestamp":1582777645797}`
	publicKeyPemFile, privateKeyPemFile := "../test/rsa/public.pem", "../test/rsa/private.pem"
	publicKeyPemBytes, err1 := ioutil.ReadFile(publicKeyPemFile)
	if err1 != nil {
		t.Fatal(err1)
	}
	privateKeyPemBytes, err2 := ioutil.ReadFile(privateKeyPemFile)
	if err2 != nil {
		t.Fatal(err2)
	}
	// RSA.Encrypt + base64.Encode
	publicKeyEncrypt := NewRSAPublicKeyEncrypt(publicKeyPemBytes)
	encryptedBytes, err1 := publicKeyEncrypt.EncryptPKCS1v15([]byte(origData))
	if err1 != nil {
		t.Fatal(err1)
	}
	encryptedBase64Go := base64.StdEncoding.EncodeToString(encryptedBytes)
	t.Log(origData)
	t.Log(encryptedBase64Go)
	// base64.Decode + RSA.Decrypt
	privateKeyDecrypt := NewRSAPrivateKeyDecrypt(privateKeyPemBytes)
	encrypted, _ := base64.StdEncoding.DecodeString(encryptedBase64Go)
	origDataBytes, err2 := privateKeyDecrypt.DecryptPKCS1v15(encrypted)
	if err2 != nil {
		t.Fatal(err2)
	}
	origDataGo := string(origDataBytes)
	if origData != origDataGo {
		t.Log(origDataGo)
		t.Fatal(" origData != origDataGo ")
	}
}

func TestCryptoAesCBCEncryptWithHmacSHA1_ABP(t *testing.T) {
	key, salt, iv := []byte("TmIhgugCGFpU7S3v"), []byte("hgt!16kl"), []byte("jkE49230Tf093b42")
	origData := []byte(`{"Customer":"gbxy","SecretIdCard":"9c1c0dd59ff33f9ac37bd072ac2df86d","Timestamp":1582777645797}`)
	encryptedBytes := CryptoAesCBCEncryptWithHmacSHA1(origData, key, salt, iv, 1000, 32)
	encryptedBase64 := `dpMVVfOuahSQ/e3o9OCKcplkG756+0R7bWMeW931IHHHNbJM9Hif80s80Wt9CMmDK81fN1JpTqiiMmLRtmLo5tzdoGyXIkinSVNokXHNw4HAC5oHljXWs3JKm6W2+D8H`
	encryptedBase64Go := base64.StdEncoding.EncodeToString(encryptedBytes)
	if encryptedBase64 != encryptedBase64Go {
		t.Log(encryptedBase64)
		t.Log(encryptedBase64Go)
		t.Fatal(" encryptedBase64 != encryptedBase64Go ")
	}
}
