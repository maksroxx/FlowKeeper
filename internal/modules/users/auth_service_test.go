package users

import (
	"errors"
	"testing"

	"github.com/maksroxx/flowkeeper/internal/config"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	Repository
	usersDB map[string]*User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		usersDB: make(map[string]*User),
	}
}

func (m *mockUserRepository) CreateUser(user *User) error {
	if _, exists := m.usersDB[user.Email]; exists {
		return errors.New("user already exists")
	}
	user.ID = uint(len(m.usersDB) + 1)
	m.usersDB[user.Email] = user
	return nil
}

func (m *mockUserRepository) GetUserByEmail(email string) (*User, error) {
	user, exists := m.usersDB[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func TestAuthService_RegisterUser(t *testing.T) {
	mockRepo := newMockUserRepository()
	cfg := config.AuthConfig{JWTSecret: "test-secret", TokenTTLHours: 24}
	svc := NewAuthService(mockRepo, cfg)

	t.Run("Успешная_регистрация", func(t *testing.T) {
		password := "secret123"
		user, err := svc.RegisterUser("Ivan", "ivan@test.com", password, 1)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Ivan", user.Name)
		assert.Equal(t, "ivan@test.com", user.Email)

		assert.NotEqual(t, password, user.Password)

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		assert.NoError(t, err, "Хэш пароля должен совпадать с оригиналом")
	})
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := newMockUserRepository()
	cfg := config.AuthConfig{JWTSecret: "test-secret", TokenTTLHours: 24}
	svc := NewAuthService(mockRepo, cfg)

	password := "mypassword"
	_, _ = svc.RegisterUser("Test User", "test@test.com", password, 1)

	t.Run("Успешный_логин", func(t *testing.T) {
		token, user, err := svc.Login("test@test.com", password)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, token)
	})

	t.Run("Неверный_пароль", func(t *testing.T) {
		token, user, err := svc.Login("test@test.com", "wrongpass")
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Nil(t, user)
		assert.Empty(t, token)
	})
}
