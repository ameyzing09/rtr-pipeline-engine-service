# 7-Layer Validation Flow Implementation

This document describes the complete 7-layer validation flow implemented in the RTR Pipeline Engine Service, matching the NestJS pattern from `rtr-job-application-service`.

---

## Overview

The service implements a comprehensive multi-layer validation approach inspired by NestJS best practices:

```
Request
   ↓
┌─────────────────────────────────────────────────────────────────┐
│ Layer 3: Rate Limiter                                           │
│ - Prevents abuse with IP + Tenant rate limiting                 │
│ - Returns 429 Too Many Requests if limit exceeded               │
└─────────────────────────────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────────────────────────────┐
│ Layer 4: Strict JSON Binding                                    │
│ - Rejects unknown/extra fields in request body                  │
│ - Returns 400 Bad Request if unknown fields detected            │
│ - Prevents malicious field injection                            │
└─────────────────────────────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────────────────────────────┐
│ Layer 1: JWT Authentication (Part 1)                            │
│ - Extracts and validates JWT token from Authorization header    │
│ - Validates token signature using JWT_SECRET                    │
│ - Injects user context (user_id, tenant_id, role, email)       │
│ - Returns 401 Unauthorized if token invalid/missing/expired     │
└─────────────────────────────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────────────────────────────┐
│ Layer 1: Tenant Security (Part 2)                               │
│ - Validates x-tenant-id header is present                       │
│ - Cross-checks JWT.tenant_id === X-Tenant-Id header             │
│ - Returns 403 Forbidden if tenant ID mismatch                   │
│ - Prevents unauthorized cross-tenant access                     │
└─────────────────────────────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────────────────────────────┐
│ Layer 5: DTO Validation (Gin binding)                           │
│ - Validates required fields with binding:"required"             │
│ - Validates string lengths: min, max                            │
│ - Validates enums: oneof                                        │
│ - Validates UUIDs: uuid                                         │
│ - Nested struct validation: dive                                │
│ - Returns 400 Bad Request if validation fails                   │
└─────────────────────────────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────────────────────────────┐
│ Layer 6: Custom Business Logic Validators                       │
│ - Tenant-specific schema validation (JSON schema)               │
│ - Date range validation (publishAt < expireAt)                  │
│ - Future/past date validation                                   │
│ - Pipeline name uniqueness (database check)                     │
│ - Stage sequence validation                                     │
│ - Returns 400 Bad Request if business logic fails               │
└─────────────────────────────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────────────────────────────┐
│ Layer 7: Role-Based Access Control (RBAC)                       │
│ - Checks user role: ADMIN, HR, INTERVIEWER                      │
│ - Enforces role-based permissions per route                     │
│ - ADMIN/HR: Create and assign pipelines                         │
│ - INTERVIEWER: Read-only (GET only)                             │
│ - Returns 403 Forbidden if insufficient role                    │
└─────────────────────────────────────────────────────────────────┘
   ↓
┌─────────────────────────────────────────────────────────────────┐
│ Handler/Business Logic                                          │
│ - Execute validated request                                     │
│ - Database operations                                           │
│ - Business logic implementation                                 │
└─────────────────────────────────────────────────────────────────┘
   ↓
Response to Client
```

---

## Layer 3: Rate Limiting

**File:** `internal/middleware/rate_limiter.go`

**Purpose:** Prevent abuse and DDoS attacks by limiting requests per user/IP

**Implementation:**
- IP-based rate limiting using `ulule/limiter`
- Tenant-aware limiting (combines IP + Tenant ID)
- In-memory store for request tracking
- Configurable limit via environment variable

**Configuration:**
```bash
RATE_LIMIT_PER_MINUTE=60  # 60 requests per minute per IP:Tenant
```

**How it works:**
1. Extract client IP from request (checks X-Forwarded-For, X-Real-IP, RemoteAddr)
2. Combine IP + Tenant ID as unique key
3. Check if request count exceeds limit for this minute
4. If exceeded, return 429 Too Many Requests
5. Set X-RateLimit-* response headers

**Example Error Response:**
```json
{
  "code": "RATE_LIMIT_EXCEEDED",
  "message": "Rate limit exceeded",
  "status_code": 429,
  "details": "Too many requests. Limit: 60 requests per minute"
}
```

**Response Headers:**
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1705421460
```

---

## Layer 4: Strict JSON Binding

**File:** `internal/middleware/strict_json.go`

**Purpose:** Prevent malicious field injection by rejecting unknown fields

**Implementation:**
- Custom middleware using Go's `json.Decoder`
- `DisallowUnknownFields()` flag enabled
- Applies to POST, PUT, PATCH requests only

**How it works:**
1. Only processes POST, PUT, PATCH with application/json
2. Reads request body
3. Uses `json.Decoder` with `DisallowUnknownFields()`
4. If unknown fields detected, returns 400 Bad Request
5. Restores body for downstream handlers

**Example Request (Valid):**
```json
{
  "name": "Senior Engineer Pipeline",
  "description": "...",
  "stages": [...]
}
```

**Example Request (Invalid - Extra Field):**
```json
{
  "name": "Senior Engineer Pipeline",
  "description": "...",
  "stages": [...],
  "malicious_field": "value"  // ❌ Rejected
}
```

**Example Error Response:**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid request format",
  "status_code": 400,
  "details": "json: unknown field \"malicious_field\""
}
```

---

## Layer 5: DTO Validation

**File:** `internal/dto/pipeline_dto.go`

**Purpose:** Validate request data types, formats, and constraints

**Implementation:**
- Gin validation tags using `go-playground/validator`
- Built into Gin's binding system
- Applied automatically to all DTOs

**Validation Tags Used:**

| Tag | Purpose | Example |
|-----|---------|---------|
| `required` | Field must be present and non-empty | `binding:"required"` |
| `omitempty` | Field is optional (skip other validators if empty) | `binding:"omitempty,max=1000"` |
| `min` | Minimum string length or number | `binding:"min=3"` |
| `max` | Maximum string length or number | `binding:"max=255"` |
| `oneof` | Field must be one of specified values | `binding:"oneof=phone technical offer"` |
| `uuid` | Field must be valid UUID format | `binding:"uuid"` |
| `dive` | Validate nested structs in arrays | `binding:"required,dive"` |

**DTO Example:**
```go
type CreatePipelineDTO struct {
    // Required, 3-255 characters
    Name string `json:"name" binding:"required,min=3,max=255"`

    // Optional, max 1000 characters
    Description string `json:"description" binding:"omitempty,max=1000"`

    // Required array of 1-10 stages
    Stages []Stage `json:"stages" binding:"required,min=1,max=10,dive"`
}

type Stage struct {
    // Required, 1-100 characters
    Stage string `json:"stage" binding:"required,min=1,max=100"`

    // Must be one of the enum values
    Type string `json:"type" binding:"required,oneof=phone technical system_design behavioral offer"`
}
```

**Example Error Response:**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid request payload",
  "status_code": 400,
  "details": "Key: 'CreatePipelineDTO.Name' Error:Field validation for 'Name' failed on the 'min' tag"
}
```

---

## Layer 6: Custom Business Logic Validators

**File:** `internal/validators/validators.go`

**Purpose:** Validate business logic constraints beyond data types

**Custom Validators Implemented:**

### 1. Tenant Schema Compliance Validator
Validates custom fields against tenant-specific JSON schema
```go
ValidateTenantSchemaCompliance(tenantID, extraFields, schema)
```

**How it works:**
1. Fetch tenant's JSON schema from database
2. Use `gojsonschema` to validate fields
3. Return detailed validation errors

**Example:**
```json
{
  "extra": {
    "salary_range": "100k-150k",
    "experience_years": "five"  // ❌ Should be number
  }
}

// Error:
// "extra/experience_years: must be number"
```

### 2. Date Range Validator
Validates that start date is before end date
```go
ValidateDateRange(startDate, endDate)
```

### 3. Stage Sequence Validator
Validates logical order of pipeline stages
```go
ValidateStageSequence(stageTypes []string)
```

### 4. Pipeline Name Validator
Validates pipeline name uniqueness
```go
ValidatePipelineName(tenantID, name)
```

**Where Custom Validators Are Applied:**
- In handlers via explicit calls to validator functions
- Can be extended to use Gin's custom validators

**Example Implementation in Handler:**
```go
func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
    var body CreatePipelineDTO
    if err := c.ShouldBindJSON(&body); err != nil {
        // Layer 5 DTO validation error
        return
    }

    // Layer 6: Custom validators
    if err := validators.ValidatePipelineStages(body.Stages); err != nil {
        appErr := errors.NewAppError(
            errors.ErrValidationError,
            "Invalid stages configuration",
            err.Error(),
        )
        c.JSON(appErr.StatusCode, appErr)
        return
    }

    // Proceed to business logic
    ...
}
```

---

## Layer 7: Role-Based Access Control (RBAC)

**File:** `internal/middleware/rbac_middleware.go`

**Purpose:** Enforce role-based permissions

**Available Roles:**
- `ADMIN` - Full access to all operations
- `HR` - Can create and assign pipelines
- `INTERVIEWER` - Read-only access (GET only)

**RBAC Middleware Functions:**

```go
// Enforce specific roles
middleware.RequireRoles(context.RoleAdmin, context.RoleHR)

// Shortcuts
middleware.RequireAdmin()
middleware.RequireAdminOrHR()

// Allow read-only for INTERVIEWER
middleware.AllowReadOnly()
```

**Route Implementation:**
```go
// POST /pipeline - ADMIN/HR only
pipeline.POST("/",
    middleware.RequireRoles(context.RoleAdmin, context.RoleHR),
    h.CreatePipeline)

// GET /pipeline - All roles, INTERVIEWER read-only
pipeline.GET("/",
    middleware.AllowReadOnly(),
    h.ListPipelines)

// POST /pipeline/assign - ADMIN/HR only
pipeline.POST("/assign",
    middleware.RequireRoles(context.RoleAdmin, context.RoleHR),
    h.AssignPipeline)
```

**Error Response (Insufficient Role):**
```json
{
  "code": "INSUFFICIENT_ROLE",
  "message": "Insufficient permissions for this operation",
  "status_code": 403,
  "details": "Required roles: ADMIN, HR"
}
```

---

## Error Status Codes Reference

| Status | Error Code | Layer | Description |
|--------|-----------|-------|-------------|
| **429** | RATE_LIMIT_EXCEEDED | 3 | Too many requests |
| **400** | VALIDATION_ERROR | 4, 5, 6 | Invalid JSON/validation/business logic |
| **401** | MISSING_TOKEN | 1 | Authorization header missing |
| **401** | INVALID_TOKEN | 1 | JWT signature invalid |
| **401** | EXPIRED_TOKEN | 1 | JWT token expired |
| **401** | MISSING_TENANT_ID | 1 | x-tenant-id header missing |
| **403** | INVALID_TENANT_ID | 1 | Tenant ID mismatch |
| **403** | INSUFFICIENT_ROLE | 7 | User lacks required role |

---

## Complete Request Example

### 1. Valid Request (All Layers Pass)

**Request:**
```bash
curl -X POST http://localhost:8081/pipeline \
  -H "Authorization: Bearer eyJhbG..." \
  -H "x-tenant-id: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Senior Engineer Pipeline",
    "description": "Interview pipeline for senior engineering positions",
    "stages": [
      {"stage": "Phone Screen", "type": "phone"},
      {"stage": "Technical Interview", "type": "technical"}
    ]
  }'
```

**Validation Flow:**
- ✅ Layer 3: IP:Tenant rate limit OK (45/60 remaining)
- ✅ Layer 4: All fields known (name, description, stages)
- ✅ Layer 1a: JWT valid, contains required claims
- ✅ Layer 1b: x-tenant-id matches JWT tenant_id
- ✅ Layer 5: All DTOs valid (lengths, enums, types)
- ✅ Layer 6: Pipeline name passes business logic
- ✅ Layer 7: User is HR role (allowed)
- ✅ Handler: Create pipeline

**Response:**
```json
{
  "id": "pipeline-uuid",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Senior Engineer Pipeline",
  "stages": "[...]",
  "created_at": "2025-01-15T10:30:00Z"
}
```

### 2. Request with Unknown Field (Layer 4 Fails)

**Request:**
```bash
curl -X POST http://localhost:8081/pipeline \
  -H "Authorization: Bearer eyJhbG..." \
  -H "x-tenant-id: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pipeline",
    "description": "...",
    "stages": [...],
    "malicious_field": "value"  # ❌ Unknown field
  }'
```

**Response (400):**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid request format",
  "status_code": 400,
  "details": "json: unknown field \"malicious_field\""
}
```

### 3. Request with Invalid DTO (Layer 5 Fails)

**Request:**
```bash
{
  "name": "PI",  # ❌ Too short (min 3)
  "stages": []   # ❌ Empty (min 1)
}
```

**Response (400):**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid request payload",
  "status_code": 400,
  "details": "Key: 'CreatePipelineDTO.Name' Error:Field validation for 'Name' failed on the 'min' tag"
}
```

### 4. Rate Limit Exceeded (Layer 3 Fails)

**Response (429) after 60+ requests/minute:**
```json
{
  "code": "RATE_LIMIT_EXCEEDED",
  "message": "Rate limit exceeded",
  "status_code": 429,
  "details": "Too many requests. Limit: 60 requests per minute"
}
```

---

## Testing the Layers

### Layer 3: Rate Limiting
```bash
# Make 61 requests in 60 seconds - last one should get 429
for i in {1..61}; do
  curl -X GET http://localhost:8081/pipeline \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "x-tenant-id: $TENANT_ID"
done
```

### Layer 4: Strict JSON
```bash
# Send extra field
curl -X POST http://localhost:8081/pipeline \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -d '{"name": "Test", "stages": [], "extra_field": "value"}'
# Should get 400 - property extra_field should not exist
```

### Layer 5: DTO Validation
```bash
# Send invalid stage type
curl -X POST http://localhost:8081/pipeline \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -d '{"name": "Test", "stages": [{"stage": "Test", "type": "invalid_type"}]}'
# Should get 400 - must be one of [phone, technical, ...]
```

### Layer 7: RBAC
```bash
# INTERVIEWER trying to POST (should be read-only)
curl -X POST http://localhost:8081/pipeline \
  -H "Authorization: Bearer $JWT_TOKEN_INTERVIEWER" \
  -H "x-tenant-id: $TENANT_ID" \
  -d '{"name": "Test", "stages": [...]}'
# Should get 403 - INTERVIEWER role has read-only access
```

---

## Environment Configuration

**Required:**
```bash
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=recrutr-db

JWT_SECRET=your-super-secret-jwt-key-minimum-32-chars
```

**Optional:**
```bash
# Default: 60 requests per minute per IP:Tenant
RATE_LIMIT_PER_MINUTE=60
```

---

## Summary

| Layer | Name | File | Status Code | Purpose |
|-------|------|------|-------------|---------|
| 3 | Rate Limiting | `rate_limiter.go` | 429 | Prevent abuse |
| 4 | Strict JSON | `strict_json.go` | 400 | Reject unknown fields |
| 1 | JWT Auth | `jwt_middleware.go` | 401 | Authenticate user |
| 1 | Tenant Check | `tenant_middleware.go` | 403 | Validate tenant |
| 5 | DTO Validation | `pipeline_dto.go` | 400 | Validate data types |
| 6 | Custom Validators | `validators.go` | 400 | Validate business logic |
| 7 | RBAC | `rbac_middleware.go` | 403 | Authorize by role |

This 7-layer approach provides comprehensive protection against:
- ✅ DDoS/abuse attacks (Layer 3)
- ✅ Malicious field injection (Layer 4)
- ✅ Unauthorized access (Layer 1, 7)
- ✅ Tenant isolation breaches (Layer 1)
- ✅ Invalid data (Layer 5)
- ✅ Business logic violations (Layer 6)

