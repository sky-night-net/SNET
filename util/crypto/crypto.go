// Package crypto provides password hashing and verification utilities.
package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPasswordAsBcrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
