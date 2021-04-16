package models

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	Username string
	Password string
	State bool
	LastLogin time.Time
}

func Hash(password string) (string, error)  {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(hashPassword string, password string) error  {
	return bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
}

