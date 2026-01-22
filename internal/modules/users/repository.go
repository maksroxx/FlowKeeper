package users

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	CreateUser(user *User) error
	GetUsers() ([]User, error)
	GetUserByID(id uint) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id uint) error

	CreateRole(role *Role) error
	GetRoles() ([]Role, error)
	GetRoleByID(id uint) (*Role, error)
	UpdateRole(role *Role) error
	DeleteRole(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(user *User) error {
	return r.db.Create(user).Error
}

func (r *repository) GetUsers() ([]User, error) {
	var users []User
	err := r.db.Preload("Role").Find(&users).Error
	return users, err
}

func (r *repository) GetUserByID(id uint) (*User, error) {
	var user User
	err := r.db.Preload("Role").First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *repository) GetUserByEmail(email string) (*User, error) {
	var user User
	err := r.db.Preload("Role").Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *repository) UpdateUser(user *User) error {
	return r.db.Save(user).Error
}

func (r *repository) DeleteUser(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

func (r *repository) CreateRole(role *Role) error {
	return r.db.Create(role).Error
}

func (r *repository) GetRoles() ([]Role, error) {
	var roles []Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *repository) GetRoleByID(id uint) (*Role, error) {
	var role Role
	err := r.db.First(&role, id).Error
	return &role, err
}

func (r *repository) UpdateRole(role *Role) error {
	return r.db.Save(role).Error
}

func (r *repository) DeleteRole(id uint) error {
	return r.db.Delete(&Role{}, id).Error
}
