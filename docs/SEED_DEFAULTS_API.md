# Seed Default Pipelines API

## Overview

The `/pipeline/seed-defaults` endpoint allows authorized users to provision predefined pipeline templates for tenants. This is typically used during tenant onboarding to provide ready-to-use hiring pipelines.

## Authentication & Authorization

**All requests require:**
- `Authorization: Bearer <JWT>` - Valid JWT token
- `X-Tenant-Id: <uuid>` - Target tenant ID (must be valid UUID format)

### Authorization Rules

| Role | Allowed Operations |
|------|-------------------|
| **ADMIN** | Can seed defaults for their own tenant only (`X-Tenant-Id == JWT.tid`) |
| **HR** | Can seed defaults for their own tenant only (`X-Tenant-Id == JWT.tid`) |
| **SUPERADMIN** | Can seed defaults for ANY tenant (cross-tenant operation) |
| **INTERVIEWER** | Not authorized (403) |

> **Security Note:** Non-SUPERADMIN users attempting to set `X-Tenant-Id` to a different tenant will receive a `403 Forbidden` error.

## Endpoint

```
POST /pipeline/seed-defaults
```

## Request Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes | Bearer JWT token with valid user credentials |
| `X-Tenant-Id` | Yes | Target tenant UUID (must match JWT.tid for non-SUPERADMIN) |
| `X-Idempotency-Key` | No | Optional idempotency key for safe retries |

## Default Pipeline Templates

### Standard Hiring Pipeline (3 stages)
- **Application Review** (Manual Review, HR, 15 min)
- **Technical Interview** (Technical, Interviewer, 60 min)
- **Final Interview** (Final, Hiring Manager, 45 min)

### Comprehensive Hiring Pipeline (5 stages)
- **Application Review** (Manual Review, HR, 15 min)
- **Phone Screening** (Phone, HR, 30 min)
- **Technical Assessment** (Technical, Interviewer, 90 min)
- **Behavioral Interview** (Behavioral, Interviewer, 60 min)
- **Final Interview** (Final, Hiring Manager, 45 min)

## Response Codes

| Code | Meaning | Description |
|------|---------|-------------|
| **201 Created** | Success - New pipelines created | At least one default pipeline was created |
| **200 OK** | Success - All exist | All default pipelines already exist (idempotent) |
| **400 Bad Request** | Invalid request | Missing/invalid `X-Tenant-Id` header |
| **401 Unauthorized** | Authentication failed | Missing, invalid, or expired JWT |
| **403 Forbidden** | Authorization failed | Role not allowed, or non-SUPERADMIN trying cross-tenant operation, or feature flag disabled |
| **404 Not Found** | Tenant not found | Target tenant does not exist |
| **409 Conflict** | Tenant not active | Tenant is in PENDING/SUSPENDED/DELETED/FAILED state |
| **500 Internal Server Error** | Server error | Database or internal service error |
| **503 Service Unavailable** | Dependency unavailable | Auth/tenant service unavailable |

## Response Body

### Success Response (201 Created)

```json
{
  "created": ["uuid-1", "uuid-2"],
  "skipped": [],
  "tenant_id": "tenant-uuid",
  "seeded_by": "user-uuid",
  "message": "Successfully created default pipelines"
}
```

### Idempotent Response (200 OK)

```json
{
  "created": [],
  "skipped": ["Standard Hiring Pipeline", "Comprehensive Hiring Pipeline"],
  "tenant_id": "tenant-uuid",
  "seeded_by": "user-uuid",
  "message": "All default pipelines already exist"
}
```

### Error Response

```json
{
  "error": "Error message",
  "code": "ERROR_CODE"
}
```

## Example Requests

### Self-Seed (Tenant Admin)

```bash
curl -X POST https://api.example.com/pipeline/seed-defaults \
  -H "Authorization: Bearer <admin-jwt>" \
  -H "X-Tenant-Id: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json"
```

### Cross-Tenant Seed (SUPERADMIN)

```bash
curl -X POST https://api.example.com/pipeline/seed-defaults \
  -H "Authorization: Bearer <superadmin-jwt>" \
  -H "X-Tenant-Id: 660e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json"
```

### With Idempotency Key

```bash
curl -X POST https://api.example.com/pipeline/seed-defaults \
  -H "Authorization: Bearer <admin-jwt>" \
  -H "X-Tenant-Id: 550e8400-e29b-41d4-a716-446655440000" \
  -H "X-Idempotency-Key: seed-tenant-550e8400-20250124" \
  -H "Content-Type: application/json"
```

## Idempotency

The endpoint is idempotent:
- Multiple calls with the same tenant will NOT create duplicate pipelines
- Returns `200 OK` if all templates already exist
- Returns `201 Created` only if new pipelines are created
- Unique constraint enforced on `(tenant_id, name)`

### Recommended Retry Strategy

```
Initial request → 5xx error
  ↓
Wait with exponential backoff: 100ms, 200ms, 400ms
  ↓
Retry up to 3 times
  ↓
Treat 200/201 as success
```

## Feature Flag

This endpoint is controlled by the `PIPELINE_DEFAULTS_ENABLED` environment variable:

```env
PIPELINE_DEFAULTS_ENABLED=true  # Enable (default)
PIPELINE_DEFAULTS_ENABLED=false # Disable (returns 403)
```

When disabled, all requests return:

```json
{
  "error": "Feature not enabled",
  "feature": "PIPELINE_DEFAULTS_ENABLED",
  "code": "FEATURE_DISABLED"
}
```

## Audit Trail

Every seed operation is logged with:
- `request_id` - Unique request identifier
- `requested_by_uid` - User ID from JWT (`created_by` in database)
- `requested_by_role` - User role (ADMIN/HR/SUPERADMIN)
- `requested_by_tenant` - User's tenant ID from JWT
- `effective_tenant` - Target tenant ID (`tenant_id` in database)
- `created` - Number of pipelines created
- `skipped` - Number of pipelines skipped (already exist)
- `status` - HTTP status code (200/201)

## Integration Flow (Tenant Onboarding)

```
1. User creates new tenant via auth-service
   POST /tenant/create → 201 Created {id: "tenant-uuid"}

2. Auth-service calls pipeline-service (server-to-server)
   POST /pipeline/seed-defaults
   Headers:
     Authorization: Bearer <superadmin-jwt>
     X-Tenant-Id: tenant-uuid

3. Pipeline-service validates and seeds
   - Validates JWT
   - Checks feature flag
   - Validates tenant state (ACTIVE)
   - Creates default pipelines

4. Response
   - 201: Pipelines created successfully
   - 200: Pipelines already exist (idempotent retry)
   - 5xx: Retry with exponential backoff
```

## Best Practices

1. **Always provide X-Idempotency-Key** for server-to-server calls
2. **Retry on 5xx errors** with exponential backoff
3. **Treat 200 and 201 as success** (both are valid idempotent responses)
4. **Log request_id** for debugging and support
5. **Monitor 403 errors** (may indicate authorization issues)
6. **Don't retry 4xx errors** (except network failures) - fix the request instead

## Security Considerations

1. **Header Validation**: `X-Tenant-Id` must be valid UUID format
2. **Cross-Check Enforcement**: Non-SUPERADMIN users cannot access other tenants
3. **Tenant State Validation**: Only ACTIVE tenants can be seeded
4. **Feature Flag**: Can disable feature globally for maintenance
5. **Audit Logging**: All operations logged for compliance

## Monitoring Metrics

Track these metrics for operational health:

- `seed_defaults_requests_total{status, role, cross_tenant}`
- `seed_defaults_pipelines_created_total`
- `seed_defaults_duration_seconds`
- `seed_defaults_errors_total{code}`
- `tenant_validation_failures_total`

## Troubleshooting

### 403 Forbidden - Tenant Mismatch

**Problem:** ADMIN/HR trying to seed different tenant

```json
{
  "error": "Tenant ID mismatch - non-SUPERADMIN users can only access their own tenant"
}
```

**Solution:** Ensure `X-Tenant-Id` matches the tenant ID in the JWT

### 403 Forbidden - Feature Disabled

**Problem:** Feature flag is off

```json
{
  "error": "Feature not enabled",
  "feature": "PIPELINE_DEFAULTS_ENABLED"
}
```

**Solution:** Set `PIPELINE_DEFAULTS_ENABLED=true` in environment

### 400 Bad Request - Invalid Tenant ID

**Problem:** `X-Tenant-Id` is not a valid UUID

```json
{
  "error": "X-Tenant-Id must be a valid UUID",
  "code": "INVALID_TENANT_ID_FORMAT"
}
```

**Solution:** Provide valid UUID format: `550e8400-e29b-41d4-a716-446655440000`

### 409 Conflict - Tenant Not Active

**Problem:** Target tenant is SUSPENDED/DELETED/PENDING

```json
{
  "error": "Tenant is not in ACTIVE state",
  "code": "TENANT_NOT_ACTIVE",
  "status": "SUSPENDED"
}
```

**Solution:** Activate the tenant first, or contact support
