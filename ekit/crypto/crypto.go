// Package crypto 加密相关
package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

// ------------------使用x/crypto/bcrypt库----------------

// 加密：生成密码哈希
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost = 10，范围 4-31，越大越安全但越慢
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// 解密（验证）：比对明文和哈希
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
