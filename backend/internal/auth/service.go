package auth

import (
	"errors"

	"url-shortener/internal/user"
	"url-shortener/internal/utils"
)

type Service interface {
	Register(username, password string) error
	Login(username, password string) (*LoginResponse, error)
}

type service struct {
	userRepo user.Repository
}

func NewService(userRepo user.Repository) Service {
	return &service{userRepo: userRepo}
}

// JWT service (nên đưa vào env sau)
var JWTService = utils.NewJWTService("secret-key")

// ===== Response =====
type LoginResponse struct {
	Token  string `json:"token"`
	UserID int64  `json:"userId"`
}

// ===== Service methods =====

func (s *service) Register(username, password string) error {
	exists, _ := s.userRepo.GetByUsername(username)
	if exists != nil {
		return errors.New("username already exists")
	}

	hashed := utils.HashPassword(password)
	return s.userRepo.Create(username, hashed)
}

func (s *service) Login(username, password string) (*LoginResponse, error) {
	u, err := s.userRepo.GetByUsername(username)
	if err != nil || u == nil {
		return nil, errors.New("invalid username or password")
	}

	if !utils.CheckPasswordHash(password, u.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}

	token, err := JWTService.Generate(int64(u.ID))
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:  token,
		UserID: int64(u.ID),
	}, nil
}
