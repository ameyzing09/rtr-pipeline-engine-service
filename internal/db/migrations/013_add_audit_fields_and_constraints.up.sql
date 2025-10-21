-- Add audit fields to pipelines table
ALTER TABLE pipelines
    ADD COLUMN created_by CHAR(36) COMMENT 'Logical FK to users.id in auth service',
    ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;

-- Add audit fields to pipeline_assignments table
ALTER TABLE pipeline_assignments
    ADD COLUMN assigned_by CHAR(36) COMMENT 'Logical FK to users.id in auth service',
    ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;

-- Add unique constraint: one pipeline name per tenant
ALTER TABLE pipelines
    ADD UNIQUE INDEX idx_tenant_pipeline_name (tenant_id, name);

-- Add unique constraint: one assignment per job per tenant
ALTER TABLE pipeline_assignments
    ADD UNIQUE INDEX idx_tenant_job_assignment (tenant_id, job_id);

-- Add composite index for querying pipelines by tenant and creator
ALTER TABLE pipelines
    ADD INDEX idx_tenant_created_by (tenant_id, created_by);

-- Add composite index for querying assignments by tenant and pipeline
ALTER TABLE pipeline_assignments
    ADD INDEX idx_tenant_pipeline (tenant_id, pipeline_id);
