package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID        int64          `gorm:"primaryKey"`
	UUID      uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null"`
	Name      string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	CreatedAt time.Time      `gorm:"type:timestamp(0)"`
	UpdatedAt time.Time      `gorm:"type:timestamp(0)"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamp(0);index"`
}

type User struct {
	ID          int64          `gorm:"primaryKey"`
	UUID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null"`
	RoleID      string         `gorm:"column:role_id;type:uuid;not null;index"`
	Username    string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password    string         `gorm:"type:text;not null"`
	Active      bool           `gorm:"not null;default:true"`
	LoginFailed int            `gorm:"not null;default:0"`
	LastLoginAt *time.Time     `gorm:"type:timestamp(0)"`
	CreatedAt   time.Time      `gorm:"type:timestamp(0)"`
	UpdatedAt   time.Time      `gorm:"type:timestamp(0)"`
	DeletedAt   gorm.DeletedAt `gorm:"type:timestamp(0);index"`

	Role Role `gorm:"foreignKey:RoleID;references:UUID"`
}

func (Role) TableName() string {
	return "roles"
}

func (User) TableName() string {
	return "users"
}

type UserToken struct {
	UserName    string     `gorm:"column:user_name"`
	UserUUID    string     `gorm:"column:user_uuid"`
	RoleName    string     `gorm:"column:role_name"`
	RoleUUID    string     `gorm:"column:role_uuid"`
	Active      bool       `gorm:"column:active"`
	LastLoginAt *time.Time `gorm:"column:last_login_at"`
}

type UserSummary struct {
	Total               int
	UUID                string
	Username            string
	RoleID              int
	Role                string
	LastLogin           *time.Time
	LastSeen            *time.Time
	Active              bool
	LastActivity        *time.Time
	LastActivityType    string
	LastActivityMessage *string
}
