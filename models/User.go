package models

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID        string    `bson:"id"`
	Username  string    `bson:"username"`
	Password  string    `bson:"password"`
	State     bool      `bson:"state"`
	LastLogin time.Time `bson:"last_login"`
}

func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(hashPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
}
