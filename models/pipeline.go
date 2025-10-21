package models

import (
	"time"

	"gorm.io/datatypes"
)

type Pipeline struct {
	ID          string         `gorm:"primaryKey;type:char(36);"`
	TenantID    string         `gorm:"type:char(36);not null;index"`
	Name        string         `gorm:"type:varchar(255);not null;uniqueIndex:idx_tenant_pipeline_name"`
	Description string         `gorm:"type:text"`
	Stages      datatypes.JSON `gorm:"type:json;not null"`
	IsActive    bool           `gorm:"default:true"`
	IsDeleted   bool           `gorm:"default:false"`
	CreatedBy   string         `gorm:"type:char(36);index:idx_tenant_created_by"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
}

type PipelineAssignment struct {
	ID         string    `gorm:"primaryKey;type:char(36);"`
	TenantID   string    `gorm:"type:char(36);not null;index:idx_tenant_pipeline"`
	PipelineID string    `gorm:"type:char(36);not null;index:idx_tenant_pipeline"`
	JobID      string    `gorm:"type:char(36);not null;index;uniqueIndex:idx_tenant_job_assignment"`
	IsDeleted  bool      `gorm:"default:false"`
	AssignedBy string    `gorm:"type:char(36)"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}
