package e2db

import (
	"time"

	"gorm.io/gorm"
)

type ModelWithSoftDelete struct {
	PKAID     uint           `gorm:"primarykey;column:pkaid" json:"pkaid"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Model struct {
	PKAID     uint      `gorm:"primarykey;column:pkaid" json:"pkaid"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `gorm:"index" json:"updated_at"`
}
