package service

import (
	"context"
	"errors"

	"github.com/truongnqse05461/chat-app/models"
	"github.com/truongnqse05461/chat-app/repository"
	"github.com/truongnqse05461/chat-app/utils"
	"golang.org/x/crypto/bcrypt"
)

type User interface {
	Register(ctx context.Context, user models.User) error
	Login(ctx context.Context, username, password string) (string, error)
}

type userService struct {
	userRepository repository.User
}

// Login implements User
func (u *userService) Login(ctx context.Context, username string, password string) (string, error) {
	user, err := u.userRepository.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if !comparePwd(user.Password, password) {
		return "", errors.New("password incorrect")
	}
	token, err := utils.CreateToken(user.Username)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Register implements User
func (u *userService) Register(ctx context.Context, user models.User) error {
	panic("unimplemented")
}

func comparePwd(hash, pass string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass)) == nil
}

var _ User = (*userService)(nil)
