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
