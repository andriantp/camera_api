package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ProtocolONVIF = "ONVIF"
	ProtocolISAPI = "ISAPI"
)

type Protocol struct {
	ID        int64          `gorm:"primaryKey"`
	UUID      uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null"`
	Name      string         `gorm:"type:varchar(50);uniqueIndex;not null"`
	CreatedAt time.Time      `gorm:"type:timestamp(0)"`
	UpdatedAt time.Time      `gorm:"type:timestamp(0)"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamp(0);index"`
}
