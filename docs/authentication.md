# Authentication

## Overview

Custom JWT-based authentication. Simple access token with 1-hour expiry, no refresh token.

## Flow

```
1. Register:  POST /api/v1/auth/register { email, password, display_name }
              → 201 { profile }

2. Login:     POST /api/v1/auth/login { email, password }
              → 200 { access_token, expires_in: 3600 }

3. Use token: GET /api/v1/events
              Header: Authorization: Bearer <access_token>

4. Token expired (after 1 hour):
              → 401 { error: "invalid or expired token" }
              → Login again
```

## Password Security

| Setting | Value |
|---------|-------|
| Algorithm | bcrypt |
| Cost factor | 12 |
| Storage | `password_hash` field in `users` collection |
| JSON serialization | Excluded (`json:"-"`) — never sent to client |

Implementation: `golang.org/x/crypto/bcrypt`

## JWT Token

| Setting | Value |
|---------|-------|
| Algorithm | HS256 (HMAC-SHA256) |
| Secret | `JWT_SECRET` env var |
| Expiry | 1 hour |
| Refresh | None — login again |

### Token Payload (Claims)

```json
{
  "sub": "665f1a2b3c4d5e6f7a8b9c0d",
  "email": "user@example.com",
  "iat": 1700000000,
  "exp": 1700003600
}
```

| Claim | Description |
|-------|-------------|
| `sub` | User's MongoDB ObjectID (hex string) |
| `email` | User's email |
| `iat` | Issued at (Unix timestamp) |
| `exp` | Expires at (Unix timestamp, iat + 3600) |

Implementation: `internal/common/jwt/jwt.go`

## Auth Middleware

The `Auth(secret)` middleware function:

1. Reads `Authorization` header
2. Validates `Bearer <token>` format
3. Parses and validates JWT (signature + expiry)
4. Stores `userID` and `email` in Gin context
5. Calls `c.Next()` on success, `c.Abort()` on failure

### Extracting User in Handlers

```go
userID := middleware.UserIDFromContext(c)  // returns string (ObjectID hex)
email := middleware.EmailFromContext(c)    // returns string
```

Implementation: `internal/common/middleware/auth.go`

## Protected vs Public Routes

| Route | Auth Required |
|-------|--------------|
| `POST /api/v1/auth/register` | No |
| `POST /api/v1/auth/login` | No |
| `GET /health` | No |
| Everything else | **Yes** |

## Error Responses

| Scenario | HTTP Status | Error Code |
|----------|-------------|------------|
| No Authorization header | 401 | UNAUTHORIZED |
| Invalid header format (not "Bearer xxx") | 401 | UNAUTHORIZED |
| Invalid/expired JWT | 401 | UNAUTHORIZED |
| Email already registered | 409 | CONFLICT |
| Wrong email or password | 401 | UNAUTHORIZED |

Login intentionally returns generic "invalid email or password" for both wrong email and wrong password (security best practice — prevents email enumeration).

## Key Files

| File | Purpose |
|------|---------|
| `internal/common/jwt/jwt.go` | GenerateToken, ParseToken |
| `internal/common/middleware/auth.go` | Auth middleware, UserIDFromContext |
| `internal/user/service.go` | Register (bcrypt), Login (JWT generation) |
| `internal/user/handler.go` | HTTP handlers |
| `internal/user/routes.go` | Public + protected route registration |
