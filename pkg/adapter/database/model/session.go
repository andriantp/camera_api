package model

import (
	"time"
)

type Session struct {
	ID         int64      `gorm:"primaryKey"`
	UUID       string     `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null"`
	UserID     string     `gorm:"column:user_id;type:uuid;not null;index"`
	LoginAt    time.Time  `gorm:"type:timestamp(0)"`
	LogoutAt   *time.Time `gorm:"type:timestamp(0)"`
	LastSeenAt time.Time  `gorm:"type:timestamp(0)"`
	IPAddress  string     `gorm:"type:inet"`
	UserAgent  string     `gorm:"type:text"`
	CreatedAt  time.Time  `gorm:"type:timestamp(0)"`
	UpdatedAt  time.Time  `gorm:"type:timestamp(0)"`

	User User `gorm:"foreignKey:UserID;references:UUID"`
}

func (Session) TableName() string {
	return "sessions"
}

type UserSession struct {
	UserName    string     `gorm:"column:user_name"`
	UserUUID    string     `gorm:"column:user_uuid"`
	Password    string     `gorm:"column:password"`
	RoleName    string     `gorm:"column:role_name"`
	RoleUUID    string     `gorm:"column:role_uuid"`
	Active      bool       `gorm:"column:active"`
	LastLoginAt *time.Time `gorm:"column:last_login_at"`
	LoginFailed int        `gorm:"column:login_failed"`
}
