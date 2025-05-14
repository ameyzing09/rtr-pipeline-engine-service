-- +migrate Down
DROP INDEX IF EXISTS idx_candidate_stage_progress_stage ON candidate_stage_progress;
DROP INDEX IF EXISTS idx_candidate_stage_progress_app ON candidate_stage_progress;

DROP TABLE IF EXISTS stage_feedback;
DROP TABLE IF EXISTS candidate_stage_progress;
DROP TABLE IF EXISTS pipeline_assignments;
DROP TABLE IF EXISTS pipelines;
DROP TABLE IF EXISTS tenants;
