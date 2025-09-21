-- +migrate Down
ALTER TABLE `stage_feedback`
    DROP FOREIGN KEY `fk_stage_feedback_tenants`,
    DROP COLUMN `tenant_id`;

ALTER TABLE `candidate_stage_progress`
    DROP FOREIGN KEY `fk_candidate_stage_progress_tenants`,
    DROP COLUMN `tenant_id`;

ALTER TABLE `pipeline_assignments`
    DROP FOREIGN KEY `fk_pipeline_assignments_tenants`,
    DROP COLUMN `tenant_id`;

ALTER TABLE `pipelines`
    DROP FOREIGN KEY `fk_pipelines_tenants`,
    DROP COLUMN `tenant_id`;
