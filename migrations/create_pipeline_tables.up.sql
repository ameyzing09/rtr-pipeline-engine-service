-- +migrate Up
CREATE TABLE tenants (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pipelines (
    id CHAR(36) PRIMARY KEY,
    tenant_id CHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    stages JSON NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE TABLE pipeline_assignments (
    id CHAR(36) PRIMARY KEY,
    tenant_id CHAR(36) NOT NULL,
    pipeline_id CHAR(36) NOT NULL,
    job_id CHAR(36) NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (pipeline_id) REFERENCES pipelines(id)
);

CREATE TABLE candidate_stage_progress (
    id CHAR(36) PRIMARY KEY,
    tenant_id CHAR(36) NOT NULL,
    application_id CHAR(36) NOT NULL,
    pipeline_id CHAR(36) NOT NULL,
    current_stage VARCHAR(255) NOT NULL,
    stage_index INT NOT NULL,
    updated_by CHAR(36),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (pipeline_id) REFERENCES pipelines(id)
);

CREATE INDEX idx_candidate_stage_progress_app ON candidate_stage_progress(application_id);
CREATE INDEX idx_candidate_stage_progress_stage ON candidate_stage_progress(stage_index);

CREATE TABLE stage_feedback (
    id CHAR(36) PRIMARY KEY,
    tenant_id CHAR(36) NOT NULL,
    application_id CHAR(36) NOT NULL,
    stage_name VARCHAR(255),
    given_by CHAR(36),
    feedback TEXT,
    score FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
