# RTR Pipeline Engine Service - API Overview

## Authentication & Authorization

### Required Headers

All API requests require:

```
Authorization: Bearer <JWT>
X-Tenant-Id: <tenant-uuid>
```

### Single Header Policy

**Important:** Only `X-Tenant-Id` header is used for tenant identification. There is no separate `X-Target-Tenant-Id` header.

- **ADMIN/HR**: `X-Tenant-Id` must match `JWT.tid` (enforced by middleware)
- **SUPERADMIN**: `X-Tenant-Id` can be any valid tenant UUID (cross-tenant operations)
- **INTERVIEWER**: `X-Tenant-Id` must match `JWT.tid`, read-only access

## Role Permission Matrix

| Endpoint | ADMIN | HR | INTERVIEWER | SUPERADMIN | Notes |
|----------|-------|----|----|------------|-------|
| `POST /pipeline` | ✅ | ✅ | ❌ | ✅ | Create pipeline |
| `GET /pipeline` | ✅ | ✅ | ✅ (read-only) | ✅ | List pipelines |
| `GET /pipeline/:id` | ✅ | ✅ | ✅ (read-only) | ✅ | Get pipeline by ID |
| `PATCH /pipeline/:id` | ✅ | ✅ | ❌ | ✅ | Update pipeline |
| `GET /pipeline/assignment?job_id=...` | ✅ | ✅ | ✅ (read-only) | ✅ | Get assignment by job |
| `POST /pipeline/assign` | ✅ | ✅ | ❌ | ✅ | Assign pipeline to job |
| `POST /pipeline/seed-defaults` | ✅ (own) | ✅ (own) | ❌ | ✅ (any) | Seed default templates |

### Cross-Tenant Rules

| Role | Cross-Tenant Allowed | Restrictions |
|------|---------------------|--------------|
| **ADMIN** | ❌ No | Can only access their own tenant (`X-Tenant-Id == JWT.tid`) |
| **HR** | ❌ No | Can only access their own tenant (`X-Tenant-Id == JWT.tid`) |
| **INTERVIEWER** | ❌ No | Can only access their own tenant, read-only |
| **SUPERADMIN** | ✅ Yes | Can access any tenant via `X-Tenant-Id` header |

**Example Cross-Tenant Operation:**
```bash
# SUPERADMIN seeding defaults for a different tenant
curl -X POST https://api.example.com/pipeline/seed-defaults \
  -H "Authorization: Bearer <superadmin-jwt>" \
  -H "X-Tenant-Id: tenant-456"   # Can be different from JWT.tid
```

## API Endpoints

### Health & Monitoring

#### GET /
Basic health check (no auth required)

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-24T10:00:00Z"
}
```

#### GET /health
Same as GET / (no auth required)

#### GET /health/detailed
Detailed health with metrics and build info (no auth required)

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-24T10:00:00Z",
  "build": {
    "version": "1.0.0",
    "build_time": "2025-01-24T09:00:00Z",
    "git_commit": "abc123",
    "go_version": "go1.21.5"
  },
  "runtime": {
    "uptime_seconds": 3600,
    "uptime_human": "1h0m0s",
    "num_goroutines": 15,
    "memory_mb": 45
  },
  "environment": "production",
  "metrics": {
    "total_requests": 1250,
    "auth_errors": 12,
    "business_errors": 8,
    "server_errors": 0,
    "requests_by_status": {
      "2xx": 1200,
      "4xx": 50,
      "5xx": 0
    },
    "latency_ms": {
      "min": 5,
      "max": 250,
      "avg": 45
    },
    "error_breakdown": {
      "401_403": 12,
      "404_409": 8,
      "5xx": 0
    }
  }
}
```

### Pipeline Management

#### POST /pipeline
Create a new pipeline

**Auth:** ADMIN, HR, SUPERADMIN
**Headers:** Authorization, X-Tenant-Id

**Request:**
```json
{
  "name": "Engineering Interview Pipeline",
  "description": "Pipeline for technical roles",
  "stages": [
    {
      "stage": "Phone Screen",
      "type": "phone",
      "conducted_by": "hr",
      "metadata": {
        "duration_minutes": 30
      }
    },
    {
      "stage": "Technical Interview",
      "type": "technical",
      "conducted_by": "interviewer",
      "metadata": {
        "duration_minutes": 60
      }
    }
  ]
}
```

**Responses:**
- `200 OK` - Pipeline created
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Missing/invalid JWT
- `403 Forbidden` - Insufficient permissions
- `409 Conflict` - Pipeline name already exists for tenant

#### GET /pipeline
List all pipelines for tenant

**Auth:** All roles (INTERVIEWER read-only)
**Headers:** Authorization, X-Tenant-Id

**Response:**
```json
[
  {
    "id": "uuid-1",
    "tenant_id": "tenant-uuid",
    "name": "Standard Pipeline",
    "description": "...",
    "stages": [...],
    "is_active": true,
    "created_by": "user-uuid",
    "created_at": "2025-01-20T10:00:00Z"
  }
]
```

#### GET /pipeline/:id
Get pipeline by ID

**Auth:** All roles (INTERVIEWER read-only)
**Headers:** Authorization, X-Tenant-Id

**Response:**
```json
{
  "id": "uuid-1",
  "tenant_id": "tenant-uuid",
  "name": "Standard Pipeline",
  "description": "...",
  "stages": [...],
  "is_active": true,
  "created_by": "user-uuid",
  "created_at": "2025-01-20T10:00:00Z"
}
```

**Errors:**
- `404 Not Found` - Pipeline not found

#### PATCH /pipeline/:id
Update pipeline

**Auth:** ADMIN, HR, SUPERADMIN
**Headers:** Authorization, X-Tenant-Id

**Request:**
```json
{
  "name": "Updated Pipeline Name",
  "description": "Updated description",
  "stages": [...]
}
```

**Responses:**
- `200 OK` - Pipeline updated
- `404 Not Found` - Pipeline not found
- `409 Conflict` - Name conflict with another pipeline

### Pipeline Assignments

#### GET /pipeline/assignment?job_id=<uuid>
Get pipeline assignment for a job

**Auth:** All roles (INTERVIEWER read-only)
**Headers:** Authorization, X-Tenant-Id
**Query Params:** `job_id` (required)

**Response (200 OK):**
```json
{
  "pipeline_id": "pipeline-uuid",
  "job_id": "job-uuid",
  "assigned_at": "2025-01-20T10:00:00Z"
}
```

**Errors:**
- `400 Bad Request` - Missing job_id query parameter
- `404 Not Found` - No assignment found for this job

#### POST /pipeline/assign
Assign pipeline to job (idempotent)

**Auth:** ADMIN, HR, SUPERADMIN
**Headers:** Authorization, X-Tenant-Id

**Request:**
```json
{
  "pipeline_id": "pipeline-uuid",
  "job_id": "job-uuid"
}
```

**Responses:**

**201 Created** - New assignment created:
```json
{
  "message": "Pipeline assigned successfully",
  "pipeline_id": "pipeline-uuid",
  "job_id": "job-uuid"
}
```

**200 OK** - Assignment already exists (idempotent):
```json
{
  "message": "Pipeline assignment already exists (idempotent)",
  "pipeline_id": "pipeline-uuid",
  "job_id": "job-uuid"
}
```

**Errors:**
- `404 Not Found` - Pipeline not found or doesn't belong to tenant
- `409 Conflict` - Job already assigned to a different pipeline

### Default Pipeline Templates

#### POST /pipeline/seed-defaults
Seed default pipeline templates for a tenant

**Auth:** ADMIN (own), HR (own), SUPERADMIN (any)
**Headers:** Authorization, X-Tenant-Id, X-Idempotency-Key (optional)

**Self-Seed Example (ADMIN/HR):**
```bash
curl -X POST https://api.example.com/pipeline/seed-defaults \
  -H "Authorization: Bearer <admin-jwt>" \
  -H "X-Tenant-Id: <same-as-jwt.tid>"
```

**Cross-Tenant Example (SUPERADMIN):**
```bash
curl -X POST https://api.example.com/pipeline/seed-defaults \
  -H "Authorization: Bearer <superadmin-jwt>" \
  -H "X-Tenant-Id: <any-tenant-uuid>"
```

**Responses:**

**201 Created** - New templates created:
```json
{
  "created": ["uuid-1", "uuid-2"],
  "skipped": [],
  "tenant_id": "tenant-uuid",
  "seeded_by": "user-uuid",
  "message": "Successfully created default pipelines"
}
```

**200 OK** - Templates already exist:
```json
{
  "created": [],
  "skipped": ["Standard Hiring Pipeline", "Comprehensive Hiring Pipeline"],
  "tenant_id": "tenant-uuid",
  "seeded_by": "user-uuid",
  "message": "All default pipelines already exist"
}
```

**Templates Created:**
1. **Standard Hiring Pipeline** (3 stages)
2. **Comprehensive Hiring Pipeline** (5 stages)

See [SEED_DEFAULTS_API.md](./SEED_DEFAULTS_API.md) for complete documentation.

## Error Codes Reference

| HTTP Code | Error Code | Description |
|-----------|------------|-------------|
| 400 | `MISSING_TENANT_ID` | X-Tenant-Id header missing |
| 400 | `INVALID_TENANT_ID_FORMAT` | X-Tenant-Id not valid UUID |
| 400 | `MISSING_JOB_ID` | job_id query parameter missing |
| 401 | `UNAUTHORIZED` | Missing/invalid/expired JWT |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 403 | `FEATURE_DISABLED` | Feature flag disabled |
| 404 | `PIPELINE_NOT_FOUND` | Pipeline not found |
| 404 | `ASSIGNMENT_NOT_FOUND` | No assignment found for job |
| 409 | `TENANT_NOT_ACTIVE` | Tenant not in ACTIVE state |
| 409 | `PIPELINE_ASSIGNMENT_CONFLICT` | Job assigned to different pipeline |
| 409 | `DUPLICATE_PIPELINE` | Pipeline name conflict |

## Idempotency

The following endpoints support idempotent operations:

| Endpoint | Idempotency Behavior |
|----------|---------------------|
| `POST /pipeline/assign` | Returns 200 if same assignment exists, 409 if different pipeline |
| `POST /pipeline/seed-defaults` | Returns 200 if templates exist, 201 if new ones created |

**Best Practice:** Include `X-Idempotency-Key` header for server-to-server calls:
```bash
curl -X POST ... \
  -H "X-Idempotency-Key: unique-operation-id-123"
```

## Tenant State Caching

Tenant status validation uses a 60-second cache:
- **Cache Hit:** Immediate response
- **Cache Miss:** Validates with auth-service, then caches for 60s
- **Performance:** Reduces auth-service load by 95%+

## Observability

### Metrics Available
- Total requests
- Requests by status code (2xx, 4xx, 5xx)
- Requests by endpoint
- Error breakdown (401/403, 404/409, 5xx)
- Latency metrics (min, max, avg)

### Monitoring Endpoints
- `GET /health` - Basic health check
- `GET /health/detailed` - Full metrics and build info

### Structured Logging
All requests logged with:
- `request_id` - Unique request identifier
- `tenant_id` - Tenant from JWT
- `user_id` - User from JWT
- `role` - User role
- `target_tenant` - Effective tenant (may differ for SUPERADMIN)
- `status` - HTTP status code
- `duration` - Request duration

## Quick Reference

### Common Request Patterns

**Self-Tenant Operation (ADMIN/HR):**
```bash
curl -X <METHOD> <URL> \
  -H "Authorization: Bearer <jwt>" \
  -H "X-Tenant-Id: <same-as-jwt.tid>" \
  -H "Content-Type: application/json"
```

**Cross-Tenant Operation (SUPERADMIN):**
```bash
curl -X <METHOD> <URL> \
  -H "Authorization: Bearer <superadmin-jwt>" \
  -H "X-Tenant-Id: <any-valid-tenant-uuid>" \
  -H "Content-Type: application/json"
```

**Read-Only Operation (INTERVIEWER):**
```bash
curl -X GET <URL> \
  -H "Authorization: Bearer <interviewer-jwt>" \
  -H "X-Tenant-Id: <same-as-jwt.tid>"
```

### Testing Cross-Tenant Authorization

```bash
# This should succeed (SUPERADMIN)
curl -X POST .../seed-defaults \
  -H "Authorization: Bearer <superadmin-jwt>" \
  -H "X-Tenant-Id: tenant-different-from-jwt"

# This should fail with 403 (ADMIN)
curl -X POST .../seed-defaults \
  -H "Authorization: Bearer <admin-jwt>" \
  -H "X-Tenant-Id: tenant-different-from-jwt"
```
