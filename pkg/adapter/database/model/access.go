package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CameraAccess struct {
	ID        int64          `gorm:"primaryKey"`
	UUID      uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null"`
	UserID    string         `gorm:"column:user_id;type:uuid;not null;index"`
	CameraID  string         `gorm:"column:camera_id;type:uuid;not null;index"`
	CreatedBy string         `gorm:"type:uuid;not null;index"`
	Active    bool           `gorm:"not null;default:true"`
	CreatedAt time.Time      `gorm:"type:timestamp(0)"`
	UpdatedAt time.Time      `gorm:"type:timestamp(0)"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamp(0);index"`

	User   User   `gorm:"foreignKey:UserID;references:UUID"`
	Camera Camera `gorm:"foreignKey:CameraID;references:UUID"`
}

func (CameraAccess) TableName() string {
	return "camera_access"
}
