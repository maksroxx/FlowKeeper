package users

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetAllUsers() ([]User, error)
	GetUserByID(id uint) (*User, error)
	CreateUser(name, email, password string, roleID uint) (*User, error)
	UpdateUser(id uint, name, email string, roleID uint, password string) (*User, error)
	DeleteUser(id uint) error

	GetAllRoles() ([]Role, error)
	CreateRole(name string, permissions []string) (*Role, error)
	UpdateRole(id uint, name string, permissions []string) (*Role, error)
	DeleteRole(id uint) error
}

type userService struct {
	repo Repository
}

func NewUserService(repo Repository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetAllUsers() ([]User, error) {
	return s.repo.GetUsers()
}

func (s *userService) GetUserByID(id uint) (*User, error) {
	return s.repo.GetUserByID(id)
}

func (s *userService) CreateUser(name, email, password string, roleID uint) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Name:     name,
		Email:    email,
		Password: string(hash),
		RoleID:   roleID,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}
	return s.repo.GetUserByID(user.ID)
}

func (s *userService) UpdateUser(id uint, name, email string, roleID uint, password string) (*User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	user.Name = name
	user.Email = email
	user.RoleID = roleID

	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hash)
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}
	return s.repo.GetUserByID(id)
}

func (s *userService) DeleteUser(id uint) error {
	return s.repo.DeleteUser(id)
}

func (s *userService) GetAllRoles() ([]Role, error) {
	return s.repo.GetRoles()
}

func (s *userService) CreateRole(name string, permissions []string) (*Role, error) {
	role := &Role{
		Name:        name,
		Permissions: permissions,
	}
	if err := s.repo.CreateRole(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *userService) UpdateRole(id uint, name string, permissions []string) (*Role, error) {
	role, err := s.repo.GetRoleByID(id)
	if err != nil {
		return nil, err
	}
	role.Name = name
	role.Permissions = permissions
	if err := s.repo.UpdateRole(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *userService) DeleteRole(id uint) error {
	return s.repo.DeleteRole(id)
}
