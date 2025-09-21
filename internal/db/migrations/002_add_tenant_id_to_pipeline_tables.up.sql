-- +migrate Up
ALTER TABLE `pipelines`
    ADD COLUMN `tenant_id` CHAR(36) NOT NULL AFTER `id`,
    ADD CONSTRAINT `fk_pipelines_tenants`
        FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`id`);

ALTER TABLE `pipeline_assignments`
    ADD COLUMN `tenant_id` CHAR(36) NOT NULL AFTER `id`,
    ADD CONSTRAINT `fk_pipeline_assignments_tenants`
        FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`id`);

ALTER TABLE `candidate_stage_progress`
    ADD COLUMN `tenant_id` CHAR(36) NOT NULL AFTER `id`,
    ADD CONSTRAINT `fk_candidate_stage_progress_tenants`
        FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`id`);

ALTER TABLE `stage_feedback`
    ADD COLUMN `tenant_id` CHAR(36) NOT NULL AFTER `id`,
    ADD CONSTRAINT `fk_stage_feedback_tenants`
        FOREIGN KEY (`tenant_id`) REFERENCES `tenants`(`id`);
