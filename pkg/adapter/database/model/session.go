package model

import (
	"time"
)

type Session struct {
	ID         int64      `gorm:"primaryKey"`
	UUID       string     `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null"`
	UserID     string     `gorm:"column:user_id;type:uuid;not null;index;index:idx_sessions_user_last_seen,priority:1"`
	LoginAt    time.Time  `gorm:"type:timestamp(0)"`
	LogoutAt   *time.Time `gorm:"type:timestamp(0)"`
	LastSeenAt time.Time  `gorm:"type:timestamp(0);index:idx_sessions_user_last_seen,priority:2"`
	IPAddress  string     `gorm:"type:inet"`
	UserAgent  string     `gorm:"type:text"`
	CreatedAt  time.Time  `gorm:"type:timestamp(0)"`
	UpdatedAt  time.Time  `gorm:"type:timestamp(0)"`

	User User `gorm:"foreignKey:UserID;references:UUID"`
}

func (Session) TableName() string {
	return "sessions"
}

// ActiveSession is the result of checking that a session is still usable:
// not logged out, and its user is active and not deleted.
type ActiveSession struct {
	RoleName string `gorm:"column:role_name"`
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
	LockedUntil *time.Time `gorm:"column:locked_until"`
}
