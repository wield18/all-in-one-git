package cipherx

import (
	"fmt"
	"testing"
)

func TestAbc(*testing.T) {
	// enc, _ := encrypt("13800138000")

	coder, _ := New([]byte("32-byte-secret-key-for-aes-256!!"))
	enc := coder.Encrypt("1ddddddddddddddddddddddddddddddddd嘟嘟嘟嘟嘟嘟哒哒哒哒哒哒") // 13800138000
	fmt.Println("加密后:", enc)

	dec, _ := coder.Decrypt(enc)
	fmt.Println("解密后:", dec)
}

var hmacKey = []byte("dwdw")

func TestDecode(t *testing.T) {
	coder, _ := New([]byte("32-byte-secret-key-for-aes-256!!"))
	dec, err := coder.Decrypt("enc")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("解密后:", dec)
	}
}
