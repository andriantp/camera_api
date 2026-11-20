package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CameraAccess grants an operator or user access to one camera.
// idx_camera_access_camera_user allows one live grant per camera/user pair;
// it is partial (deleted_at IS NULL) so a soft-deleted grant doesn't block
// assigning the camera again.
type CameraAccess struct {
	ID        int64          `gorm:"primaryKey"`
	UUID      uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex;not null"`
	UserID    string         `gorm:"column:user_id;type:uuid;not null;index;uniqueIndex:idx_camera_access_camera_user,priority:2,where:deleted_at IS NULL"`
	CameraID  string         `gorm:"column:camera_id;type:uuid;not null;index;uniqueIndex:idx_camera_access_camera_user,priority:1,where:deleted_at IS NULL"`
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
