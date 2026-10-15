package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Direction struct {
	ID        int64     `gorm:"primaryKey"`
	UUID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null"`
	CreatedBy string    `gorm:"type:uuid;not null;index"`
	Name      string    `gorm:"type:varchar(100);not null"`

	LeftID  *string `gorm:"column:left_id;type:uuid;index"`
	RightID *string `gorm:"column:right_id;type:uuid;index"`
	UpID    *string `gorm:"column:up_id;type:uuid;index"`
	DownID  *string `gorm:"column:down_id;type:uuid;index"`

	CreatedAt time.Time      `gorm:"type:timestamp(0)"`
	UpdatedAt time.Time      `gorm:"type:timestamp(0)"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamp(0);index"`

	Left  *PanTilt `gorm:"foreignKey:LeftID;references:UUID"`
	Right *PanTilt `gorm:"foreignKey:RightID;references:UUID"`
	Up    *PanTilt `gorm:"foreignKey:UpID;references:UUID"`
	Down  *PanTilt `gorm:"foreignKey:DownID;references:UUID"`

	CreatedUser *User `gorm:"foreignKey:CreatedBy;references:UUID"`
}

type PanTilt struct {
	ID        int64     `gorm:"primaryKey"`
	UUID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null"`
	CreatedBy string    `gorm:"type:uuid;not null;index"`
	Name      string    `gorm:"type:varchar(100);not null"`

	Pan      float64       `gorm:"not null"`
	Tilt     float64       `gorm:"not null"`
	Duration time.Duration `gorm:"type:bigint;not null"`

	CreatedAt time.Time      `gorm:"type:timestamp(0)"`
	UpdatedAt time.Time      `gorm:"type:timestamp(0)"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamp(0);index"`

	CreatedUser *User `gorm:"foreignKey:CreatedBy;references:UUID"`
}
