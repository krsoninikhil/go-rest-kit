# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**go-rest-kit** is a Go library that provides ready-made packages for quickly setting up REST APIs using the Gin framework. The primary goal is to eliminate repetitive handler code, request parsing, and validation by leveraging Go generics to create FastAPI-like controller methods.

Key principle: Controllers receive parsed request objects as arguments and return response types, avoiding manual parsing in every handler.

## Development Commands

### Running Tests
```bash
go test ./...                    # Run all tests
go test ./auth                   # Run tests for a specific package
go test -v ./auth                # Run with verbose output
```

Note: Currently only the `auth` package has tests (`auth/otp_service_test.go`).

### Dependencies
```bash
go mod download                  # Download dependencies
go mod tidy                      # Clean up go.mod and go.sum
```

### Running the Example
```bash
cd examples
go run .                         # Runs the example API server on :8080
```

## Architecture Overview

This is a **library/framework** (not a standalone application). It provides reusable components that other projects import and use to build REST APIs.

### Layer Architecture

The framework follows a clean architecture pattern with three layers:

1. **Controller Layer** (`crud.Controller`, `crud.NestedController`): Handles HTTP requests/responses
2. **Service Layer** (`crud.Service` interface): Business logic (optional, can use DAO directly)
3. **DAO Layer** (`crud.Dao`): Database access using GORM

### Core Package Responsibilities

#### `request` - Request Binding & Response Handling
- `BindAll(handler)`: Binds URI, query params, headers, and JSON body
- `BindGet(handler)`: For GET requests (URI + query params only)
- `BindCreate(handler)`: For POST requests (URI + query + JSON body)
- `BindUpdate(handler)`: For PATCH/PUT requests (URI + query + JSON body)
- `BindDelete(handler)`: For DELETE requests (URI + query params only)
- `BindNestedCreate(handler)`: For nested resource POST requests
- `Respond(ctx, response, error)`: Centralized response handling with proper HTTP status codes

These functions convert FastAPI-style handlers `func(*gin.Context, Req) (*Res, error)` into standard Gin handlers.

#### `crud` - Generic CRUD Controllers & DAOs
**Key Interfaces:**
- `Model`: Must be implemented by all database models (provides `IsDeleted()`, `ResourceName()`, `PK()`, `Joins()`)
- `ModelWithCreator`: Optional interface for user-scoped resources (provides `SetCreatedBy()`, `CreatedByID()`)
- `Request[M]`: Converts request DTOs to models via `ToModel(*gin.Context) M`
- `Response[M]`: Converts models to response DTOs via `FillFromModel(M) Response[M]`
- `Service[M]`: Defines CRUD operations (Get, Create, Update, Delete, List, BulkCreate)

**Controllers:**
- `Controller[M, S, R]`: Standard CRUD controller for resources like `/resource/:id`
- `NestedController[M, S, R]`: For nested resources like `/parent/:parentID/resource/:id`

**DAO:**
- `Dao[M]`: Generic GORM-based DAO implementing `Service[M]` interface
- Handles pagination via `pgdb.Page` (supports cursor and offset pagination)
- Automatically filters by creator ID for models implementing `ModelWithCreator`

#### `pgdb` - PostgreSQL Database Utilities
- `BaseModel`: Embeddable struct with ID, CreatedAt, UpdatedAt, DeletedAt (soft delete support)
- `NewPGConnection(ctx, config)`: Creates GORM DB connection
- `Paginate(page, sortField)`: GORM scope for pagination (used in DAO List methods)
- Pagination supports both cursor-based (After) and offset-based (Page) approaches

#### `auth` - Authentication System
- **OTP-based authentication flow**: send OTP → verify OTP → JWT tokens
  - Uses SMS providers (Twilio, Fast2SMS) for OTP delivery
  - Supports test phone numbers for development/app review
- **Generic OAuth authentication flow**: exchange auth code → get user info → JWT tokens
  - **Extensible provider system**: Easy to add new OAuth providers (Google, Twitter, LinkedIn, etc.)
  - `OAuthProvider` interface: Common interface all providers implement
  - `NewGoogleOAuthProvider(config)`: Google OAuth implementation (more providers can be added)
  - `OAuthAuth` controller handler: Single unified endpoint for all OAuth providers
  - Requires email verification from OAuth provider
  - User model must implement `SetOAuthInfo(OAuthUserInfo)` method
  - `WithOAuthProvider(provider)`: Add OAuth provider to controller
- **JWT token generation and validation**: access + refresh tokens with configurable expiry
- `GinStdMiddleware(config)`: Gin middleware to protect routes with JWT
- `NewUserDao[T]`: Generic user DAO for any user model implementing `UserModel` interface
  - Supports both phone-based (OTP) and email-based (OAuth) user lookup
- Country info lookup for phone number validation

#### `apperrors` - Application Errors
- `AppError`: Base interface with `HTTPCode()` and `HTTPResponse()` methods
- `ServerError`: 500 errors with cause tracking
- `NotFoundError`: 404 errors
- `PermissionError`: 403 errors
- `InvalidParamsError`: 400 errors
- `ConflictError`: 409 errors (e.g., duplicate key violations)

#### `config` - Configuration Management
- Loads config from YAML files with environment variable overrides
- Pattern: Empty YAML values are filled from `.env` file (e.g., `REDIS_PASSWORD` env var fills `redis.password`)
- `Load(ctx, &config)`: Main loading function
- Implement `AppConfig` interface: `EnvPath()`, `SourcePath()`, `SetEnv(string)`

#### `integrations` - Third-Party Integrations
- `twilio`: SMS provider for OTP delivery
- `fast2sms`: Alternative SMS provider

#### `cache` - Cache Interface
- `NewInMemory()`: Simple in-memory cache implementation (for examples/testing)
- Used by OTP service to store temporary OTP codes

## Generic Type Patterns

When creating CRUD APIs for a new resource:

1. Define your model implementing `crud.Model` (embed `pgdb.BaseModel` for defaults)
2. Define request/response types implementing `crud.Request[M]` and `crud.Response[M]`
3. Create DAO: `crud.Dao[YourModel]{PGDB: db}`
4. Create Controller: `crud.Controller[Model, Response, Request]{Svc: &dao}`
5. Register routes: `r.GET("/resource/:id", request.BindGet(controller.Retrieve))`

For nested resources (e.g., products under a business):
- Model must implement `NestedModel[M]` (adds `ParentID()` and `SetParentID(int) M`)
- Use `crud.NestedController` instead of `Controller`
- Use `request.BindNestedCreate` for POST endpoints
- Route pattern: `/parent/:parentID/resource/:id`

## Important Implementation Notes

### Request Binding Order
The `request` package binds in this order: URI params → Query params → Headers → JSON body. This matters because URI params might be marked `binding:"required"` and must be checked first.

### Response Type Assertions
Controllers use type assertions to convert interface responses back to concrete types. If `FillFromModel` is implemented incorrectly (returns wrong type), it will panic with a clear message.

### Permission Checking
- Resources implementing `ModelWithCreator` automatically get filtered by user ID in List operations
- Retrieve/Update/Delete operations verify creator ID matches authenticated user
- Nested resources verify `ParentID` matches the URI parameter

### Pagination
- Cursor-based: Use `After` parameter (returns `NextAfter` in response)
- Offset-based: Use `Page` parameter
- `Limit` controls page size (applied to both strategies)
- DAO's List method uses `pgdb.Paginate` scope for sorting and pagination

### Joins/Preloads
Models specify related data to load via `Joins()` method returning field names for GORM's `Preload()`. These are automatically applied in DAO's Get and List methods.

## OAuth Integration (Google, Twitter, LinkedIn, etc.)

The auth package supports multiple OAuth providers through a generic, extensible design:

1. **Add config to your YAML**:
```yaml
auth:
  oauth:
    google:
      client_id: "your-google-client-id"
      client_secret: "your-google-client-secret"
      redirect_url: "http://localhost:8080/auth/google/callback"
    # Add more providers as needed
    # twitter:
    #   client_id: "your-twitter-client-id"
    #   client_secret: "your-twitter-client-secret"
    #   redirect_url: "http://localhost:8080/auth/twitter/callback"
```

2. **Update your User model** to implement `SetOAuthInfo`:
```go
func (u User) SetOAuthInfo(info auth.OAuthUserInfo) auth.UserModel {
    u.Email = info.Email
    u.Name = info.Name
    u.Picture = info.Picture
    u.OAuthID = info.ProviderID
    u.OAuthProvider = info.Provider  // "google", "twitter", etc.
    if info.Locale != "" {
        u.Locale = info.Locale
    }
    return u
}
```

3. **Initialize OAuth providers** and add to controller:
```go
// Create providers for enabled OAuth methods
googleProvider := auth.NewGoogleOAuthProvider(conf.Auth.OAuth["google"])
// twitterProvider := auth.NewTwitterOAuthProvider(conf.Auth.OAuth["twitter"])

// Add to controller
authController := auth.NewController(authSvc, otpSvc, cache).
    WithOAuthProvider(googleProvider)
    // .WithOAuthProvider(twitterProvider)
```

4. **Register the route**:
```go
r.POST("/auth/oauth", request.BindCreate(authController.OAuthAuth))
```

The frontend should:
1. Complete OAuth flow with provider to get authorization code
2. Send code and provider name to `/auth/oauth`: `{"code": "...", "provider": "google"}`
3. Receive JWT tokens (same format as OTP flow)

**Adding New Providers**: Implement `OAuthProvider` interface in a new file (see `google_oauth_provider.go` as reference), add to factory in `oauth_provider.go`, and update validation in `entities.go`.

## Example Code Reference

See `examples/main.go` for a complete working example showing:
- Configuration loading
- Database connection setup
- Auth service initialization (OTP + JWT)
- Standard CRUD resource setup (BusinessType)
- User-scoped resource setup (Business)
- Nested resource setup (Products under Business)
- Route registration with appropriate middleware
