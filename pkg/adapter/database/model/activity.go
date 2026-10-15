package model

import (
	"time"

	"github.com/google/uuid"
)

type UserActivity struct {
	ID           int64     `gorm:"primaryKey"`
	UUID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null"`
	ActorUserID  string    `gorm:"column:actor_user_id;type:uuid;not null;index"`
	TargetUserID *string   `gorm:"column:target_user_id;type:uuid;index"`
	CameraID     *string   `gorm:"column:camera_id;type:uuid;index"`
	SessionID    *string   `gorm:"column:session_id;type:uuid;index"`
	IPAddress    string    `gorm:"type:inet"`
	UserAgent    string    `gorm:"type:text"`
	Activity     string    `gorm:"type:varchar(100);not null;index"`
	Message      *string   `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"type:timestamp(0)"`

	ActorUser  User     `gorm:"foreignKey:ActorUserID;references:UUID"`
	TargetUser *User    `gorm:"foreignKey:TargetUserID;references:UUID"`
	Camera     *Camera  `gorm:"foreignKey:CameraID;references:UUID"`
	Session    *Session `gorm:"foreignKey:SessionID;references:UUID"`
}

type UserActivitySummary struct {
	Total     int
	UUID      string
	At        time.Time
	Actor     *string
	IPAddress string
	UserAgent string
	Activity  string
	Target    *string
	Camera    *string
	Message   *string
}
