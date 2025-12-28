package auth

import (
	"errors"
	"url-shortener/internal/user"
	"url-shortener/internal/utils"
)

type Service interface {
	Register(username, password string) error
	Login(username, password string) (string, error)
}

type service struct {
	userRepo user.Repository
}

func NewService(userRepo user.Repository) Service {
	return &service{userRepo: userRepo}
}

var JWTService = utils.NewJWTService("secret-key") 

func (s *service) Register(username, password string) error {
	exists, _ := s.userRepo.GetByUsername(username)
	if exists != nil {
		return errors.New("username already exists")
	}
	hashed := utils.HashPassword(password)
	return s.userRepo.Create(username, hashed)
}

func (s *service) Login(username, password string) (string, error) {
	u, err := s.userRepo.GetByUsername(username)
	if err != nil || u == nil {
		return "", errors.New("invalid username or password")
	}
	if !utils.CheckPasswordHash(password, u.PasswordHash) {
		return "", errors.New("invalid username or password")
	}
	token, _ := JWTService.Generate(int64(u.ID))
	return token, nil
}
