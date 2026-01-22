package users

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type Permissions []string

func (p Permissions) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Permissions) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, p)
}

type Role struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	Name        string      `gorm:"unique;not null" json:"name"`
	Permissions Permissions `gorm:"type:text" json:"permissions"`
}

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"unique;not null" json:"name"`
	Email    string `gorm:"unique;not null" json:"email"`
	Password string `gorm:"not null" json:"-"`

	RoleID uint `json:"role_id"`
	Role   Role `json:"role"`

	gorm.Model
}
