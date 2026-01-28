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
	if value == nil {
		*p = Permissions{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("type assertion failed: value is not []byte or string")
	}

	if len(bytes) == 0 {
		*p = Permissions{}
		return nil
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
