# API Usage Guide

Quick reference for using the RTR Pipeline Engine Service API with security features.

## Quick Start

### 1. Prerequisites

- JWT token with valid claims (user_id, tenant_id, role, email)
- Tenant ID matching the JWT tenant_id
- Valid role: ADMIN, HR, or INTERVIEWER

### 2. Request Headers

**Required:**
```
Authorization: Bearer <jwt_token>
x-tenant-id: <tenant-uuid>
```

Note: The x-tenant-id must match the tenant_id claim in the JWT token, or the request will be rejected with 403 Forbidden.

### 3. Role-Based Endpoints

#### Create Pipeline (ADMIN/HR only)

```bash
curl -X POST http://localhost:8081/pipeline \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Senior Engineer Pipeline",
    "description": "Interview pipeline for senior positions",
    "stages": [
      {"stage": "Phone Screen", "type": "phone", "conducted_by": "hr"},
      {"stage": "Technical Interview", "type": "technical", "conducted_by": "interviewer"},
      {"stage": "System Design", "type": "system_design", "conducted_by": "interviewer"},
      {"stage": "Behavioral", "type": "behavioral", "conducted_by": "interviewer"},
      {"stage": "Offer", "type": "offer", "conducted_by": "auto"}
    ]
  }'
```

**Validation Rules:**
- `name`: Required, 3-255 characters
- `description`: Optional, max 1000 characters
- `stages`: Required array with at least 1 stage
  - `stage`: Required, 1-100 characters (stage name)
  - `type`: Required, 1-50 characters (stage type)
  - `conducted_by`: Required, 1-50 characters (who conducts this stage)
  - `metadata`: Optional object for custom fields

**Response (200 OK):**
```json
{
  "id": "pipeline-uuid",
  "tenant_id": "tenant-uuid",
  "name": "Senior Engineer Pipeline",
  "description": "Interview pipeline for senior positions",
  "stages": "[...]",
  "is_active": true,
  "is_deleted": false,
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Response (400 Bad Request - Validation Failed):**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid request payload",
  "status_code": 400,
  "details": "Field validation errors: Name - This field is required, Stages - Value is too short (minimum: 1)"
}
```

**Error Response (409 Conflict - Duplicate Pipeline Name):**
```json
{
  "code": "DUPLICATE_PIPELINE",
  "message": "Pipeline with this name already exists",
  "status_code": 409,
  "details": "A pipeline with this name already exists for your tenant. Please use a different name."
}
```

**Note:** Pipeline names must be unique within a tenant. The uniqueness constraint is enforced on (tenant_id, name).

#### List Pipelines (All roles, INTERVIEWER read-only)

```bash
curl -X GET http://localhost:8081/pipeline \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID"
```

**Response (200 OK):**
```json
[
  {
    "id": "pipeline-uuid-1",
    "tenant_id": "tenant-uuid",
    "name": "Senior Engineer Pipeline",
    "description": "...",
    "stages": "[...]",
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z"
  },
  {
    "id": "pipeline-uuid-2",
    "tenant_id": "tenant-uuid",
    "name": "Junior Developer Pipeline",
    "description": "...",
    "stages": "[...]",
    "is_active": true,
    "created_at": "2024-01-15T11:00:00Z"
  }
]
```

#### Get Pipeline by ID (All roles, INTERVIEWER read-only)

```bash
curl -X GET http://localhost:8081/pipeline/{pipeline-id} \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID"
```

**Response (200 OK):**
```json
{
  "id": "pipeline-uuid-1",
  "tenant_id": "tenant-uuid",
  "name": "Senior Engineer Pipeline",
  "description": "Interview pipeline for senior positions",
  "stages": "[...]",
  "is_active": true,
  "is_deleted": false,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Error Response (404 Not Found):**
```json
{
  "code": "PIPELINE_NOT_FOUND",
  "message": "Pipeline not found",
  "status_code": 404,
  "details": "The requested pipeline does not exist or you don't have access to it"
}
```

#### Update Pipeline (ADMIN/HR only)

**Partial Update Support:** You can update just the fields you want to change (name, description, or stages).

```bash
# Update only the name
curl -X PATCH http://localhost:8081/pipeline/{pipeline-id} \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Pipeline Name"
  }'

# Update only stages
curl -X PATCH http://localhost:8081/pipeline/{pipeline-id} \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "stages": [
      {"stage": "New Stage 1", "type": "technical", "conducted_by": "interviewer"},
      {"stage": "New Stage 2", "type": "behavioral", "conducted_by": "hr"}
    ]
  }'

# Update multiple fields
curl -X PATCH http://localhost:8081/pipeline/{pipeline-id} \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Senior Engineer Pipeline v2",
    "description": "Updated description",
    "stages": [
      {"stage": "Phone Screen", "type": "phone", "conducted_by": "hr"},
      {"stage": "Technical Interview", "type": "technical", "conducted_by": "interviewer"}
    ]
  }'
```

**Validation Rules:**
- `name`: Optional, if provided: 3-255 characters
- `description`: Optional, if provided: max 1000 characters
- `stages`: Optional, if provided: min 1 stage with nested validation

**Response (200 OK):**
```json
{
  "id": "pipeline-uuid-1",
  "tenant_id": "tenant-uuid",
  "name": "Updated Pipeline Name",
  "description": "Updated description",
  "stages": "[...]",
  "is_active": true,
  "is_deleted": false,
  "created_by": "user-123",
  "updated_by": "user-456",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T14:20:00Z"
}
```

**Error Responses:**
- **404 Not Found:** Pipeline doesn't exist or not accessible
- **409 Conflict:** Duplicate pipeline name
- **400 Bad Request:** Validation failed

#### Assign Pipeline (ADMIN/HR only)

```bash
curl -X POST http://localhost:8081/pipeline/assign \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_id": "pipeline-uuid",
    "job_id": "job-uuid"
  }'
```

**Response (201 Created):**
```json
{
  "message": "Pipeline assigned successfully"
}
```

---

## Standardized Error Response Format

All error responses follow a consistent JSON structure:

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message",
  "status_code": 400,
  "details": "Additional context or technical details (optional)"
}
```

**Fields:**
- `code`: Machine-readable error code (e.g., "VALIDATION_ERROR", "PIPELINE_NOT_FOUND")
- `message`: Human-readable error message
- `status_code`: HTTP status code (matches the response status)
- `details`: Additional context or technical details (omitted if empty)

---

## Error Handling

### Authentication Errors (401)

**Missing Token:**
```json
{
  "code": "MISSING_TOKEN",
  "message": "Missing authorization token",
  "status_code": 401,
  "details": "Authorization header not found or malformed"
}
```

**Invalid Token:**
```json
{
  "code": "INVALID_TOKEN",
  "message": "Failed to parse or validate token",
  "status_code": 401,
  "details": "signature is invalid"
}
```

**Missing Tenant ID:**
```json
{
  "code": "MISSING_TENANT_ID",
  "message": "Missing tenant ID header",
  "status_code": 401,
  "details": "x-tenant-id header is required"
}
```

### Authorization Errors (403)

**Tenant ID Mismatch:**
```json
{
  "code": "INVALID_TENANT_ID",
  "message": "Access denied: Tenant ID in JWT does not match X-Tenant-Id header",
  "status_code": 403
}
```

**Insufficient Role:**
```json
{
  "code": "INSUFFICIENT_ROLE",
  "message": "Insufficient permissions for this operation",
  "status_code": 403,
  "details": "Required roles: ADMIN, HR"
}
```

**INTERVIEWER POST Request:**
```json
{
  "code": "FORBIDDEN",
  "message": "INTERVIEWER role has read-only access",
  "status_code": 403,
  "details": "Only GET requests are allowed"
}
```

### Validation Errors (400)

**Invalid JSON Payload:**
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid request payload",
  "status_code": 400,
  "details": "json: cannot unmarshal string into Go value of type []dto.Stage"
}
```

---

## Testing with cURL

### Generate Test Token (Node.js)

```javascript
// token_generator.js
const jwt = require('jsonwebtoken');

const payload = {
  user_id: 'user-123',
  tenant_id: 'tenant-456',
  role: 'ADMIN',
  email: 'admin@example.com'
};

const secret = 'your-super-secret-jwt-key-change-this-in-production';
const token = jwt.sign(payload, secret, { expiresIn: '1h' });

console.log(token);
```

```bash
node token_generator.js
```

### Test Request

```bash
#!/bin/bash
JWT_TOKEN="<token-from-generator>"
TENANT_ID="tenant-456"

curl -v -X GET http://localhost:8081/pipeline \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -H "Content-Type: application/json"
```

**Note:** The tenant_id in the JWT must match the x-tenant-id header, or the request will be rejected with 403 Forbidden.

---

## Role Permissions Matrix

| Operation | ADMIN | HR | INTERVIEWER |
|-----------|-------|----|----|
| List Pipelines | ✓ | ✓ | ✓ (read-only) |
| Get Pipeline by ID | ✓ | ✓ | ✓ (read-only) |
| Create Pipeline | ✓ | ✓ | ✗ |
| Update Pipeline | ✓ | ✓ | ✗ |
| Assign Pipeline | ✓ | ✓ | ✗ |

---

## Request Logging

All HTTP requests are logged with the following information:

- **method**: HTTP method (GET, POST, etc.)
- **path**: Request path (e.g., /pipeline/123)
- **status**: HTTP response status code
- **duration**: Request processing time
- **tenant_id**: Tenant ID from JWT (or "unknown" if not authenticated)
- **user_id**: User ID from JWT (or "unknown" if not authenticated)
- **ip**: Client IP address

**Example log output:**
```
[INFO] [HTTP] method=GET path=/pipeline/123 status=200 duration=45ms tenant_id=tenant-456 user_id=user-123 ip=192.168.1.100
```

**Log Levels:**
- Set via `LOG_LEVEL` environment variable
- Options: `debug`, `info`, `warn`, `error`
- Default: `info`

---

## Environment Variables

```bash
# Database
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=recrutr-db

# JWT Configuration (Required)
JWT_SECRET=your-super-secret-jwt-key-minimum-32-chars
```

---

## Common Issues

### Issue: "Missing authorization token"
**Solution:** Add `Authorization: Bearer <token>` header

### Issue: "Tenant ID mismatch"
**Solution:** Ensure x-tenant-id header matches the tenant_id in JWT token

### Issue: "INSUFFICIENT_ROLE"
**Solution:** Verify user has ADMIN or HR role for write operations

### Issue: "Invalid token"
**Solution:** Check JWT_SECRET is correct and matches token signing secret

### Issue: "INTERVIEWER role has read-only access"
**Solution:** INTERVIEWER can only use GET requests, not POST/PUT/DELETE

---

## Response Status Codes

| Code | Meaning | Common Cause |
|------|---------|--------------|
| 200 | OK | Successful GET request |
| 201 | Created | Successful POST request |
| 400 | Bad Request | Invalid JSON or missing required fields |
| 401 | Unauthorized | Invalid/missing JWT or tenant ID |
| 403 | Forbidden | Insufficient role permissions |
| 404 | Not Found | Resource doesn't exist |
| 500 | Internal Server Error | Database or server error |

