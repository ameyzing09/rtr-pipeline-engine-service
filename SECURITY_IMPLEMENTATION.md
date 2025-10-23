# Security & Middleware Implementation

This document outlines the security features and middleware implementation for the RTR Pipeline Engine Service.

## Overview

The service implements a complete security layer with:
- **JWT (JSON Web Token) authentication** for user identification
- **Role-Based Access Control (RBAC)** for authorization
- **Tenant isolation** with JWT tenant ID cross-validation
- **Standardized error codes** with proper HTTP status codes

---

## Architecture

### Middleware Stack

The middleware is applied in the following order:

1. **JWTMiddleware** - Parses and validates JWT tokens
2. **TenantMiddleware** - Validates tenant ID header and cross-checks with JWT
3. **RBAC Middleware** - Enforces role-based access control at route level

### Error Handling

All errors use standardized error codes with appropriate HTTP status codes:
- **401 Unauthorized** - Authentication errors (missing/invalid/expired token)
- **403 Forbidden** - Authorization errors (insufficient role permissions, tenant ID mismatch)
- **400 Bad Request** - Validation errors
- **500 Internal Server Error** - Server errors

---

## Detailed Component Descriptions

### 1. JWT Authentication (`internal/middleware/jwt_middleware.go`)

**Functionality:**
- Extracts Bearer token from `Authorization` header
- Validates token signature using `JWT_SECRET`
- Extracts claims: `user_id`, `tenant_id`, `role`, `email`
- Injects user context into request for use by handlers

**Expected JWT Format:**

```json
{
  "user_id": "user-uuid",
  "tenant_id": "tenant-uuid",
  "role": "ADMIN|HR|INTERVIEWER",
  "email": "user@example.com",
  "iat": 1234567890,
  "exp": 1234571490
}
```

**Header Requirement:**
```
Authorization: Bearer <jwt_token>
```

**Error Responses:**

```json
// Missing token
{
  "code": "MISSING_TOKEN",
  "message": "Missing authorization token",
  "status_code": 401
}

// Invalid/expired token
{
  "code": "INVALID_TOKEN",
  "message": "Invalid token",
  "status_code": 401,
  "details": "unexpected signing method"
}

// Expired token
{
  "code": "EXPIRED_TOKEN",
  "message": "Token has expired",
  "status_code": 401
}
```

---

### 2. Tenant Middleware (`internal/middleware/tenant_middleware.go`)

**Functionality:**
- Validates `x-tenant-id` header is present
- **Critical:** Cross-checks that JWT.tenant_id === X-Tenant-Id header
- Sets tenant ID in context for backward compatibility
- Returns 403 Forbidden if tenant IDs don't match

**Header Requirements:**

```
x-tenant-id: <tenant-uuid>
```

**Cross-Validation Logic:**
The middleware enforces that the tenant ID from the JWT token must exactly match the X-Tenant-Id header value. This prevents unauthorized access to other tenants' data.

**Error Responses:**

```json
// Missing tenant ID header
{
  "code": "MISSING_TENANT_ID",
  "message": "Missing tenant ID header",
  "status_code": 401,
  "details": "x-tenant-id header is required"
}

// Tenant ID mismatch (JWT.tenant_id !== X-Tenant-Id)
{
  "code": "INVALID_TENANT_ID",
  "message": "Access denied: Tenant ID in JWT does not match X-Tenant-Id header",
  "status_code": 403
}
```

---

### 3. Role-Based Access Control (`internal/middleware/rbac_middleware.go`)

**Available Roles:**

- `ADMIN` - Full access to all operations
- `HR` - Can create and assign pipelines
- `INTERVIEWER` - Read-only access (GET requests only)

**RBAC Middleware Functions:**

#### `RequireRoles(...roles)`
Restricts access to specified roles only.

```go
// Only ADMIN and HR can access this route
middleware.RequireRoles(context.RoleAdmin, context.RoleHR)
```

#### `RequireAdmin()`
Shortcut for admin-only access.

```go
middleware.RequireAdmin()
```

#### `RequireAdminOrHR()`
Shortcut for ADMIN or HR roles.

```go
middleware.RequireAdminOrHR()
```

#### `AllowReadOnly()`
Allows all roles, but restricts INTERVIEWER to GET requests only.

```go
middleware.AllowReadOnly()
```

**Error Response:**

```json
{
  "code": "INSUFFICIENT_ROLE",
  "message": "Insufficient permissions for this operation",
  "status_code": 403,
  "details": "Required roles: ADMIN, HR"
}
```

---

### 4. Request Context (`internal/context/context.go`)

**Structure:**

```go
type RequestContext struct {
    UserID   string   // UUID of authenticated user
    TenantID string   // UUID of user's tenant
    Role     UserRole // ADMIN, HR, or INTERVIEWER
    Email    string   // User email
}
```

**Helper Methods:**

```go
ctx.IsAdmin()                    // Returns true if user is ADMIN
ctx.IsHR()                       // Returns true if user is HR
ctx.IsInterviewer()              // Returns true if user is INTERVIEWER
ctx.HasRole(RoleAdmin, RoleHR)  // Checks if user has any of specified roles

// Retrieve in handlers:
requestCtx, exists := context.GetRequestContext(c)
if exists {
    userID := requestCtx.UserID
    role := requestCtx.Role
}
```

---

### 5. Standardized Error Codes (`internal/errors/errors.go`)

**Error Codes:**

| Code | HTTP Status | Description |
|------|------------|-------------|
| `UNAUTHORIZED` | 401 | General authentication error |
| `MISSING_TOKEN` | 401 | Authorization header missing |
| `INVALID_TOKEN` | 401 | Token signature invalid |
| `EXPIRED_TOKEN` | 401 | Token has expired |
| `MISSING_TENANT_ID` | 401 | Tenant ID header missing |
| `INVALID_TENANT_ID` | 403 | Tenant ID mismatch with JWT |
| `FORBIDDEN` | 403 | General authorization error |
| `INSUFFICIENT_ROLE` | 403 | User role lacks permissions |
| `BAD_REQUEST` | 400 | Invalid request format |
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `NOT_FOUND` | 404 | Resource not found |
| `DATABASE_ERROR` | 500 | Database operation failed |
| `INTERNAL_ERROR` | 500 | General server error |

**Error Response Format:**

```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message",
  "status_code": 400,
  "details": "Optional detailed information"
}
```

---

## Route Protection

### Pipeline Routes

| Method | Route | Auth | Role | Description |
|--------|-------|------|------|-------------|
| POST | `/pipeline` | Required | ADMIN, HR | Create new pipeline |
| GET | `/pipeline` | Required | All | List pipelines (INTERVIEWER read-only) |
| POST | `/pipeline/assign` | Required | ADMIN, HR | Assign pipeline to job |

**Route Implementation:**

```go
// POST /pipeline - Create pipeline (ADMIN/HR only)
pipeline.POST("/",
    middleware.RequireRoles(context.RoleAdmin, context.RoleHR),
    h.CreatePipeline)

// GET /pipeline - List pipelines (all roles, INTERVIEWER read-only)
pipeline.GET("/",
    middleware.AllowReadOnly(),
    h.ListPipelines)

// POST /pipeline/assign - Assign pipeline (ADMIN/HR only)
pipeline.POST("/assign",
    middleware.RequireRoles(context.RoleAdmin, context.RoleHR),
    h.AssignPipeline)
```

---

## Environment Configuration

**Required Environment Variables:**

```bash
# Database Configuration
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=ameykode
DB_NAME=recrutr-db

# JWT Configuration (Required)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
```

**Production Recommendations:**

1. Use strong, random values for all secrets (minimum 32 characters)
2. Store secrets in a secure vault (HashiCorp Vault, AWS Secrets Manager, etc.)
3. Rotate secrets regularly
4. Use different secrets per environment (dev, staging, production)
5. Never commit real secrets to version control

---

## Client Example

### Complete Request Example

```bash
# Generate JWT token (server-side or via auth service)
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
TENANT_ID="550e8400-e29b-41d4-a716-446655440000"

# Create pipeline
curl -X POST http://localhost:8081/pipeline \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "x-tenant-id: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Senior Engineer Pipeline",
    "description": "Interview pipeline for senior engineering positions",
    "stages": [
      {"stage": "Phone Screen", "type": "phone"},
      {"stage": "Technical Interview", "type": "technical"},
      {"stage": "System Design", "type": "system_design"},
      {"stage": "Behavioral", "type": "behavioral"},
      {"stage": "Offer"}
    ]
  }'
```

---

## Implementation Checklist

- [x] JWT parsing middleware with token validation
- [x] Tenant ID validation and verification
- [x] Cross-check JWT tenant ID with header tenant ID
- [x] Role-based access control middleware
- [x] Standardized error codes (401, 403)
- [x] Request context injection (user_id, tenant_id, role)
- [x] Route-level RBAC enforcement
- [x] Standardized error response format
- [x] Environment configuration
- [x] Handler error handling updates

---

## Security Considerations

1. **JWT Secrets**: All JWT signing should use a strong secret (32+ characters)
2. **Token Expiration**: Configure appropriate token expiration times
3. **HTTPS Only**: Always use HTTPS in production to prevent token interception
4. **Tenant Isolation**: JWT tenant ID must always match X-Tenant-Id header - returns 403 Forbidden on mismatch
5. **Rate Limiting**: Consider implementing rate limiting to prevent brute force attacks
6. **Audit Logging**: Log all authentication failures and role-based access denials
7. **Token Refresh**: Implement token refresh mechanism for long-lived sessions

---

## Testing

To test the middleware, ensure your JWT token includes required fields:

```json
{
  "user_id": "test-user-id",
  "tenant_id": "test-tenant-id",
  "role": "ADMIN",
  "email": "test@example.com"
}
```

Test scenarios:
- Missing Authorization header → 401
- Invalid JWT signature → 401
- Missing x-tenant-id header → 401
- Tenant ID mismatch with JWT → 403
- INTERVIEWER attempting POST → 403
- HR attempting to create pipeline → Success (200)
- ADMIN attempting any operation → Success (200)

---

## Files Modified/Created

### New Files
- `internal/errors/errors.go` - Standardized error types and codes
- `internal/context/context.go` - Request context structures
- `internal/middleware/jwt_middleware.go` - JWT token parsing
- `internal/middleware/rbac_middleware.go` - Role-based access control

### Modified Files
- `internal/middleware/tenant_middleware.go` - Validates tenant ID header and cross-checks with JWT
- `internal/routes/pipeline_routes.go` - Added RBAC middleware to routes
- `internal/handler/pipeline_handler.go` - Updated error handling
- `cmd/main.go` - Updated middleware stack
- `go.mod` - Added JWT dependency
- `.env` - Added JWT configuration

