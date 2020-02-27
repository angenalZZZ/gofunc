package f

import (
	"encoding/hex"
	"testing"
)

func TestEncryptDecryptAes128(t *testing.T) {
	origData := []byte("hello")
	key := []byte("TmIhgugCGFpU7S3v")
	iv := []byte("jkE49230Tf093b42")
	encrypted, err2 := EncryptAes128(origData, key, iv)
	if err2 != nil {
		t.Fatal(err2)
	}
	encryptedString := hex.EncodeToString(encrypted)
	// Output: EncryptAes128: hello => 548e8841b4baa92451bc4e7fd875ad1c
	t.Logf("EncryptAes128: %s => %s", origData, encryptedString)
	encryptedRaw, err3 := hex.DecodeString(encryptedString)
	if err3 != nil {
		t.Fatal(err3)
	}
	if string(encryptedRaw) != string(encrypted) {
		t.Fatal("encryptedRaw != encrypted")
	}
	origDataRaw, err4 := DecryptAes128(encryptedRaw, key, iv)
	if err4 != nil {
		t.Fatal(err4)
	}
	if string(origDataRaw) != string(origData) {
		t.Fatal("origDataRaw != origData")
	}
	// Output: DecryptAes128: 548e8841b4baa92451bc4e7fd875ad1c => hello
	t.Logf("DecryptAes128: %s => %s", hex.EncodeToString(encryptedRaw), origDataRaw)
	// pbkdf2 Rfc2898DeriveBytes password
	salt := []byte("hgt!16kl")
	encrypted1, err1 := EncryptAes128s(origData, key, salt, iv)
	if err1 != nil {
		t.Fatal(err1)
	}
	encryptedString1 := hex.EncodeToString(encrypted1)
	// Output: EncryptAes128s: hello => 5dbfefb3b68651d1e965ddcd62afd3b5
	t.Logf("EncryptAes128s: %s => %s", origData, encryptedString1)
	encryptedRaw1, err31 := hex.DecodeString(encryptedString1)
	if err31 != nil {
		t.Fatal(err31)
	}
	if string(encryptedRaw1) != string(encrypted1) {
		t.Fatal("encryptedRaw1 != encrypted1")
	}
	origDataRaw1, err41 := DecryptAes128s(encryptedRaw1, key, salt, iv)
	if err41 != nil {
		t.Fatal(err41)
	}
	if string(origDataRaw1) != string(origData) {
		t.Fatal("origDataRaw1 != origData")
	}
	// Output: DecryptAes128s: 5dbfefb3b68651d1e965ddcd62afd3b5 => hello
	t.Logf("DecryptAes128s: %s => %s", hex.EncodeToString(encryptedRaw1), origDataRaw1)
}

func TestEncryptDecryptDes128(t *testing.T) {
	origData := []byte("hello")
	key := []byte("GFpU7S3v")
	iv := []byte("jkE49230Tf093b42")
	encrypted, err2 := EncryptDes128(origData, key, iv)
	if err2 != nil {
		t.Fatal(err2)
	}
	encryptedString := hex.EncodeToString(encrypted)
	// Output: EncryptDes128: hello => 898aff98549d75cb
	t.Logf("EncryptDes128: %s => %s", origData, encryptedString)
	encryptedRaw, err3 := hex.DecodeString(encryptedString)
	if err3 != nil {
		t.Fatal(err3)
	}
	if string(encryptedRaw) != string(encrypted) {
		t.Fatal("encryptedRaw != encrypted")
	}
	origDataRaw, err4 := DecryptDes128(encryptedRaw, key, iv)
	if err4 != nil {
		t.Fatal(err4)
	}
	if string(origDataRaw) != string(origData) {
		t.Fatal("origDataRaw != origData")
	}
	// Output: DecryptDes128: 898aff98549d75cb => hello
	t.Logf("DecryptDes128: %s => %s", hex.EncodeToString(encryptedRaw), origDataRaw)
}

func TestEncryptTripleDesECB(t *testing.T) {
	origData := []byte("hello")
	key := []byte("TmIhgugCGFpU7S3vGFpU7S3v")
	encrypted, err2 := EncryptTripleDesECB(origData, key)
	if err2 != nil {
		t.Fatal(err2)
	}
	encryptedString := hex.EncodeToString(encrypted)
	// Output: EncryptTripleDesECB: hello => 86f21066c5ba8c49
	t.Logf("EncryptTripleDesECB: %s => %s", origData, encryptedString)
}
