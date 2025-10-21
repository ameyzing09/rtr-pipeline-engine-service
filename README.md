# RTR Pipeline Engine Service

Pipeline management microservice for the Recrutr platform. Handles pipeline templates and job assignments.

## Features

- **Pipeline Templates**: Create and manage reusable interview pipeline templates
- **Job Assignments**: Assign pipeline templates to specific jobs
- **Multi-tenancy**: Full tenant isolation with logical foreign keys
- **Audit Trail**: Track who created pipelines and who assigned them
- **Unique Constraints**: Enforce one pipeline per job per tenant

## Tech Stack

- **Language**: Go 1.24.3
- **Framework**: Gin
- **ORM**: GORM
- **Database**: MySQL 8.0+
- **Migrations**: golang-migrate
- **Auth**: JWT-based authentication

## Prerequisites

- Go 1.24.3 or higher
- MySQL 8.0 or higher
- Make (for migration commands)

## Setup

### 1. Clone and Install Dependencies

```bash
git clone <repository-url>
cd rtr-pipeline-engine-service
go mod download
```

### 2. Configure Environment

Copy `.env.example` to `.env` and update with your values:

```env
# Database Configuration
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=recrutr-db

# Logging Configuration
LOG_LEVEL=debug

# JWT Configuration
JWT_SECRET=your_jwt_secret_key

# Rate Limiting
RATE_LIMIT_PER_MINUTE=60

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000
CORS_MAX_AGE_HOURS=12
```

### 3. Install Migration Tool

```bash
make migrate-install
```

This installs `golang-migrate` CLI to `$GOPATH/bin/migrate`. Ensure `$GOPATH/bin` is in your PATH.

### 4. Run Migrations

```bash
# Apply all migrations
make migrate-up

# Check current migration version
make migrate-version

# Rollback N steps (default 1)
make migrate-down STEPS=2

# Force to specific version (use carefully)
make migrate-force VERSION=2
```

### 5. Start the Server

**Development (with hot reload using Air):**
```bash
make dev
```

**Manual (without hot reload):**
```bash
make run
```

## Database Schema

### Pipelines Table

Stores interview pipeline templates.

| Column | Type | Description |
|--------|------|-------------|
| `id` | CHAR(36) | Primary key (UUID) |
| `tenant_id` | CHAR(36) | Tenant ID (logical FK to tenants.id in auth service) |
| `name` | VARCHAR(255) | Pipeline name (unique per tenant) |
| `description` | TEXT | Optional description |
| `stages` | JSON | Array of interview stages |
| `is_active` | BOOLEAN | Whether pipeline is active |
| `is_deleted` | BOOLEAN | Soft delete flag |
| `created_by` | CHAR(36) | User ID who created (logical FK to users.id in auth service) |
| `created_at` | TIMESTAMP | Creation timestamp |
| `updated_at` | TIMESTAMP | Last update timestamp |

**Indexes:**
- `idx_tenant_id` on (tenant_id)
- `idx_tenant_pipeline_name` UNIQUE on (tenant_id, name)
- `idx_tenant_created_by` on (tenant_id, created_by)

### Pipeline Assignments Table

Maps pipelines to jobs.

| Column | Type | Description |
|--------|------|-------------|
| `id` | CHAR(36) | Primary key (UUID) |
| `tenant_id` | CHAR(36) | Tenant ID |
| `pipeline_id` | CHAR(36) | Pipeline ID |
| `job_id` | CHAR(36) | Job ID |
| `is_deleted` | BOOLEAN | Soft delete flag |
| `assigned_by` | CHAR(36) | User ID who assigned (logical FK to users.id in auth service) |
| `created_at` | TIMESTAMP | Creation timestamp |
| `updated_at` | TIMESTAMP | Last update timestamp |

**Indexes:**
- `idx_tenant_id` on (tenant_id)
- `idx_pipeline_id` on (pipeline_id)
- `idx_job_id` on (job_id)
- `idx_tenant_job_assignment` UNIQUE on (tenant_id, job_id)
- `idx_tenant_pipeline` on (tenant_id, pipeline_id)

## API Endpoints

All endpoints require JWT authentication via `Authorization: Bearer <token>` header and `X-Tenant-Id` header.

### Create Pipeline
```http
POST /pipeline
Content-Type: application/json

{
  "name": "Software Engineer Pipeline",
  "description": "Standard pipeline for software engineering roles",
  "stages": [
    {
      "stage": "Phone Screen",
      "type": "phone",
      "conducted_by": "hr",
      "metadata": {
        "duration_minutes": 30,
        "scoring_criteria": ["communication", "experience"]
      }
    },
    {
      "stage": "Technical Interview",
      "type": "technical",
      "conducted_by": "interviewer",
      "metadata": {
        "duration_minutes": 60,
        "required_skills": ["Python", "Data Structures"],
        "difficulty": "medium"
      }
    },
    {
      "stage": "System Design",
      "type": "system_design",
      "conducted_by": "interviewer"
    },
    {
      "stage": "Behavioral Interview",
      "type": "behavioral",
      "conducted_by": "hr"
    },
    {
      "stage": "Final Offer",
      "type": "offer",
      "conducted_by": "hr"
    }
  ]
}
```

#### Stage Structure - Fully Customizable

**Pipelines are fully customizable by each tenant.** There are no restrictions on what values you use - configure your pipeline to match your hiring process exactly.

| Field | Required | Description | Examples |
|-------|----------|-------------|----------|
| `stage` | ✅ Yes | Stage display name (1-100 chars) | "Phone Screen", "Culture Fit", "Portfolio Review" |
| `type` | ✅ Yes | Stage category/type (1-50 chars) | Predefined: `phone`, `technical`, `system_design`, `behavioral`, `offer`<br>Custom: `case_study`, `portfolio_review`, `workshop` |
| `conducted_by` | ✅ Yes | Who conducts this stage (1-50 chars) | Predefined: `auto`, `hr`, `interviewer`<br>Custom: `hiring_manager`, `team_lead`, `cto`, `peer` |
| `metadata` | ❌ No | Custom fields for your workflow | `duration_minutes`, `scoring_criteria`, `required_skills`, etc. |

#### Customization Examples

**Example 1: Design Agency Pipeline**
```json
{
  "name": "Senior Designer Hiring",
  "stages": [
    {
      "stage": "Portfolio Review",
      "type": "portfolio_review",
      "conducted_by": "creative_director",
      "metadata": {
        "evaluation_criteria": ["creativity", "technical_skill", "brand_fit"],
        "portfolio_url_required": true
      }
    },
    {
      "stage": "Design Challenge",
      "type": "case_study",
      "conducted_by": "design_team",
      "metadata": {
        "duration_hours": 4,
        "tools_allowed": ["Figma", "Adobe XD"]
      }
    },
    {
      "stage": "Team Collaboration Workshop",
      "type": "workshop",
      "conducted_by": "team",
      "metadata": {
        "participants": ["designers", "product_managers"],
        "duration_minutes": 90
      }
    }
  ]
}
```

**Example 2: Sales Role Pipeline**
```json
{
  "name": "Account Executive Hiring",
  "stages": [
    {
      "stage": "Phone Screen",
      "type": "phone",
      "conducted_by": "hr"
    },
    {
      "stage": "Sales Pitch Demo",
      "type": "role_play",
      "conducted_by": "sales_manager",
      "metadata": {
        "scenario": "cold_call_enterprise",
        "duration_minutes": 30,
        "scoring": {
          "objection_handling": 30,
          "product_knowledge": 30,
          "closing_technique": 40
        }
      }
    },
    {
      "stage": "VP Interview",
      "type": "executive_interview",
      "conducted_by": "vp_sales"
    }
  ]
}
```

**Example 3: Automated Screening Pipeline**
```json
{
  "name": "High-Volume Recruiting",
  "stages": [
    {
      "stage": "AI Resume Screen",
      "type": "automated_screening",
      "conducted_by": "auto",
      "metadata": {
        "ai_model": "resume_parser_v2",
        "min_score": 70
      }
    },
    {
      "stage": "Video Interview",
      "type": "asynchronous_video",
      "conducted_by": "auto",
      "metadata": {
        "questions": ["Tell us about yourself", "Why this role?"],
        "time_limit_per_question": 120
      }
    },
    {
      "stage": "Human Review",
      "type": "manual_review",
      "conducted_by": "recruiter"
    }
  ]
}
```

#### Common Predefined Types (Recommendations, Not Required)

While you can use any values, here are commonly used types for consistency:

**Stage Types:**
- `phone` - Phone screening
- `technical` - Technical/coding interview
- `system_design` - System design interview
- `behavioral` - Behavioral interview
- `offer` - Offer stage

**Conducted By:**
- `auto` - Automated (AI/system)
- `hr` - HR team member
- `interviewer` - Technical interviewer

### List Pipelines
```http
GET /pipeline
```

### Assign Pipeline to Job
```http
POST /pipeline/assign
Content-Type: application/json

{
  "pipeline_id": "uuid-of-pipeline",
  "job_id": "uuid-of-job"
}
```

## Migrations

### Creating New Migrations

1. Create numbered migration files in `internal/db/migrations/`:

```bash
# Example: 003_add_new_field.up.sql
ALTER TABLE pipelines ADD COLUMN new_field VARCHAR(255);

# Example: 003_add_new_field.down.sql
ALTER TABLE pipelines DROP COLUMN new_field;
```

2. Apply migrations:
```bash
make migrate-up
```

### Migration File Naming Convention

- Format: `NNN_descriptive_name.up.sql` and `NNN_descriptive_name.down.sql`
- Example: `001_create_pipelines_and_assignments.up.sql`
- Numbers should be sequential (001, 002, 003, etc.)

### Current Migrations

1. **001_create_pipelines_and_assignments**: Initial schema (pipelines and pipeline_assignments tables)
2. **002_add_audit_fields_and_constraints**: Adds audit fields (created_by, assigned_by, updated_at) and unique constraints

## Architecture

### Logical Foreign Keys

This service uses **logical foreign keys** instead of physical database constraints for references to the auth service:

- `pipelines.created_by` → references `users.id` (in auth service DB)
- `pipeline_assignments.assigned_by` → references `users.id` (in auth service DB)
- `tenant_id` → references `tenants.id` (in auth service DB)

**Why?** This allows services to scale independently and prevents cross-database constraint issues when services are separated.

### Layered Architecture

```
handlers/     → HTTP request handling, JWT extraction
services/     → Business logic
repositories/ → Database operations
models/       → Data structures
middleware/   → CORS, JWT, RBAC, rate limiting
```

## Development

### Hot Reload with Air

Air provides nodemon-like hot reloading for Go:

```bash
# Install Air
make air-install

# Start with hot reload
make dev
```

See [README_AIR.md](README_AIR.md) for more details.

### Debug Logging

Set `LOG_LEVEL=debug` in `.env` to enable debug logs:

```env
LOG_LEVEL=debug
```

## Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | MySQL host | 127.0.0.1 |
| `DB_PORT` | MySQL port | 3306 |
| `DB_USER` | MySQL user | root |
| `DB_PASSWORD` | MySQL password | - |
| `DB_NAME` | Database name | recrutr-db |
| `LOG_LEVEL` | Log level (debug, info, warn, error) | info |
| `JWT_SECRET` | JWT signing secret (must match auth service) | - |
| `RATE_LIMIT_PER_MINUTE` | Max requests per minute per IP+tenant | 60 |
| `CORS_ALLOWED_ORIGINS` | Comma-separated allowed origins | http://localhost:3000 |
| `CORS_MAX_AGE_HOURS` | CORS preflight cache duration | 12 |

## License

Proprietary - Recrutr Platform
