package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Camera struct {
	ID        int64     `gorm:"primaryKey"`
	UUID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null"`
	CreatedBy string    `gorm:"type:uuid;not null;index"`

	ProtocolID  string  `gorm:"column:protocol_id;type:uuid;not null;index"`
	DirectionID *string `gorm:"column:direction_id;type:uuid;index"`

	SN       string `gorm:"type:varchar(100);index"`
	Host     string `gorm:"type:varchar(255);not null"`
	Username string `gorm:"type:text;not null"`
	Password string `gorm:"type:text;not null"`
	Active   bool   `gorm:"not null;default:true"`

	Profile *string `gorm:"type:varchar(100)"`
	Channel *int    `gorm:"type:integer"`

	SearchID *int `gorm:"column:search_id;type:integer"`
	TrackID  *int `gorm:"column:track_id;type:integer"`

	CreatedAt time.Time      `gorm:"type:timestamp(0)"`
	UpdatedAt time.Time      `gorm:"type:timestamp(0)"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamp(0);index"`

	CreatedUser *User      `gorm:"foreignKey:CreatedBy;references:UUID"`
	Protocol    Protocol   `gorm:"foreignKey:ProtocolID;references:UUID"`
	Direction   *Direction `gorm:"foreignKey:DirectionID;references:UUID"`
}

type CameraSummary struct {
	UUID     string
	SN       string
	Protocol string
	Active   bool
}
