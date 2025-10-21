-- Remove composite indexes
ALTER TABLE pipeline_assignments
    DROP INDEX idx_tenant_pipeline;

ALTER TABLE pipelines
    DROP INDEX idx_tenant_created_by;

-- Remove unique constraints
ALTER TABLE pipeline_assignments
    DROP INDEX idx_tenant_job_assignment;

ALTER TABLE pipelines
    DROP INDEX idx_tenant_pipeline_name;

-- Remove audit fields from pipeline_assignments
ALTER TABLE pipeline_assignments
    DROP COLUMN updated_at,
    DROP COLUMN assigned_by;

-- Remove audit fields from pipelines
ALTER TABLE pipelines
    DROP COLUMN updated_at,
    DROP COLUMN created_by;
