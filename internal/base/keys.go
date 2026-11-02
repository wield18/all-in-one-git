package base

import (
	"all-in-one/config"
	"all-in-one/pkg/ekit/cipherx"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func NewKeys(conf config.Keys) (hmacKey []byte, cryptoKey string, cipherDB *cipherx.Cipher) {
	var err error
	hmacKey = []byte(conf.HMAS)
	cryptoKey = conf.CryptoKey
	cipherDB, err = cipherx.New([]byte(conf.AesGcmDbKey))
	if err != nil {
		panic(err)
	}
	return
}

// hmac 哈希列, 定长32字节
func HashHMACString(str string) []byte {
	h := hmac.New(sha256.New, hmacKey)
	h.Write([]byte(str))
	return h.Sum(nil)
}

// crypto 加密：生成密码哈希
// 一般db的密码列或者长时token使用
func HashCryptoCode(str string) (string, error) {
	by := preLenCheck(fmt.Sprint(str, cryptoKey))
	// bcrypt.DefaultCost = 10，范围 4-31，越大越安全但越慢
	bytes, err := bcrypt.GenerateFromPassword(by, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// crypto 解密（验证）：比对明文和哈希
func CheckCryptoCode(hash, str string) bool {
	by := preLenCheck(fmt.Sprint(str, cryptoKey))
	err := bcrypt.CompareHashAndPassword([]byte(hash), by)
	return err == nil
}

func preLenCheck(str string) []byte {
	var by []byte
	if len(str) > 72 {
		hashed := sha256.Sum256([]byte(str))
		by = hashed[:]
	} else {
		by = []byte(str)
	}
	return by
}

// HashSha256 一般的短时的token
func HashSha256(s string) string {
	sum := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", sum)
}
