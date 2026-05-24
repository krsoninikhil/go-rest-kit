# Optional Impersonation Extension

This package adds impersonation as an **opt-in** capability for `go-rest-kit/auth`.

## Default behavior

If you do not import or wire this package:

- Base auth routes and middleware remain unchanged.
- JWT claim shape remains the same as default auth.
- No impersonation endpoints are registered.

## Wiring overview

1. Build a `PermissionChecker` implementation for your app policy.
2. Provide `UserLookup`, `SessionStore`, and optional `AuditSink` implementations.
3. Create the extension:

```go
ext := impersonation.NewExtension(
    impersonation.NewService(conf.Auth, checker, userLookup, sessionStore, auditSink),
)
```

4. Apply middleware and routes where needed:

```go
api.Use(ext.Middleware())
impersonation.RegisterRoutes(api, ext.Controller)
```

## API endpoints

- `POST /api/admin/impersonate`
- `DELETE /api/admin/impersonate`

Both routes are only available when the extension is explicitly registered.
