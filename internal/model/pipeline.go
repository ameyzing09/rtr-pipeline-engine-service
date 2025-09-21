package model

import (
	"time"

	"gorm.io/datatypes"
)

type Pipeline struct {
	ID          string         `gorm:"primaryKey;type:char(36);"`
	TenantID    string         `gorm:"type:char(36);not null;index"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	Stages      datatypes.JSON `gorm:"type:json:not null"`
	IsActice    bool           `gorm:"default:true"`
	IsDeleted   bool           `gorm:"default:false"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
}

type PipelineAssignment struct {
	ID         string    `gorm:"primaryKey;type:char(36);"`
	TenantID   string    `gorm:"type:char(36);not null;index"`
	PipelineID string    `gorm:"type:char(36);not null;index"`
	JobID      string    `gorm:"type:char(36);not null;index"`
	IsDeleted  bool      `gorm:"default:false"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}
// /* Rectangle 1 */

// position: absolute;
// width: 720px;
// height: 72px;

// background: rgba(255, 252, 252, 0.1);
// box-shadow: 0px 8px 13px rgba(0, 0, 0, 0.25), inset 4px 5px 6px rgba(0, 0, 0, 0.4), inset -1px -3px 4px rgba(255, 255, 255, 0.4);
// backdrop-filter: blur(3.75px);
// /* Note: backdrop-filter has minimal browser support */
// border-radius: 100px;