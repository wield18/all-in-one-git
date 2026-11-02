package cipherx

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Cipher 持有可复用的 AEAD，避免每次加解密都重建。
type Cipher struct {
	aead cipher.AEAD
}

// New 用给定密钥创建 Cipher，密钥必须是 16/24/32 字节。
func New(key []byte) (*Cipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{aead: aead}, nil
}

// Encrypt 加密并 base64 编码，nonce 拼在密文前。
func (c *Cipher) Encrypt(plaintext string) string {
	nonce := make([]byte, c.aead.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	// if ; err != nil {
	// 	return "", err
	// }
	ciphertext := c.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// Decrypt 解密 base64 字符串。
func (c *Cipher) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	ns := c.aead.NonceSize()
	if len(data) < ns {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:ns], data[ns:]
	plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
