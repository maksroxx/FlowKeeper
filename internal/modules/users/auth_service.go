package users

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/maksroxx/flowkeeper/internal/config"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenGeneration    = errors.New("failed to generate token")
)

type Service interface {
	RegisterUser(name, email, password string, roleID uint) (*User, error)
	Login(email, password string) (string, *User, error)
}

type authService struct {
	repo   Repository
	config config.AuthConfig
}

func NewAuthService(repo Repository, cfg config.AuthConfig) Service {
	return &authService{
		repo:   repo,
		config: cfg,
	}
}

func (s *authService) RegisterUser(name, email, password string, roleID uint) (*User, error) {
	hashedPass, err := s.hashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Name:     name,
		Email:    email,
		Password: hashedPass,
		RoleID:   roleID,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(email, password string) (string, *User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if !s.checkPasswordHash(password, user.Password) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return "", nil, ErrTokenGeneration
	}

	return token, user, nil
}

func (s *authService) generateJWT(user *User) (string, error) {
	now := time.Now()
	ttl := time.Duration(s.config.TokenTTLHours) * time.Hour

	claims := jwt.MapClaims{
		"sub": user.ID,
		"iss": "flowkeeper",
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),

		"email":       user.Email,
		"role":        user.Role.Name,
		"permissions": user.Role.Permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *authService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *authService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
