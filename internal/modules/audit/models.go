package audit

import (
	"time"

	"github.com/maksroxx/flowkeeper/internal/modules/users"
)

type AuditLog struct {
	ID     uint       `gorm:"primaryKey" json:"id"`
	UserID uint       `gorm:"index" json:"user_id"`
	User   users.User `gorm:"constraint:OnDelete:SET NULL;" json:"user"`

	Action    string `json:"action"`
	Entity    string `json:"entity"`
	EntityID  uint   `json:"entity_id"`
	Details   string `json:"details"`
	IPAddress string `json:"ip_address"`

	CreatedAt time.Time `json:"created_at"`
}
