# Auth Package

The auth package provides authentication mechanisms for your REST API, including OTP-based phone authentication and Google OAuth.

## Features

- **OTP Authentication**: Phone number-based authentication with SMS OTP delivery
- **OAuth Authentication**: Generic OAuth provider support (Google, Twitter, LinkedIn, etc.)
  - Extensible design - easy to add new providers
  - Single unified endpoint for all OAuth providers
- **JWT Tokens**: Access and refresh token generation and validation
- **User DAO**: Generic user data access with support for both phone and email lookup

## Usage

### OTP Authentication

```go
// Setup
userDao := auth.NewUserDao[User](db)
authSvc := auth.NewService(conf.Auth, userDao)
smsProvider := twilio.NewClient(conf.Auth.Twilio)
otpSvc := auth.NewOTPSvc(conf.Auth.OTP, smsProvider, cache)
authController := auth.NewController(authSvc, otpSvc, cache)

// Register routes
r.POST("/auth/otp/send", request.BindCreate(authController.SendOTP))
r.POST("/auth/otp/verify", request.BindCreate(authController.VerifyOTP))
```

**Flow:**
1. Frontend sends phone number to `/auth/otp/send`
2. User receives SMS with OTP code
3. Frontend sends phone number + OTP to `/auth/otp/verify`
4. Backend returns JWT access and refresh tokens

### OAuth Authentication (Google, Twitter, LinkedIn, etc.)

```go
// Setup (in addition to OTP setup above)
// Initialize OAuth providers
googleProvider := auth.NewGoogleOAuthProvider(conf.Auth.OAuth["google"])
// twitterProvider := auth.NewTwitterOAuthProvider(conf.Auth.OAuth["twitter"])  // When implemented
// linkedinProvider := auth.NewLinkedInOAuthProvider(conf.Auth.OAuth["linkedin"])  // When implemented

// Add providers to controller
authController := auth.NewController(authSvc, otpSvc, cache).
	WithOAuthProvider(googleProvider)
	// .WithOAuthProvider(twitterProvider)
	// .WithOAuthProvider(linkedinProvider)

// Or add multiple at once
// authController := auth.NewController(authSvc, otpSvc, cache).
//     WithOAuthProviders(googleProvider, twitterProvider, linkedinProvider)

// Register single unified OAuth route
r.POST("/auth/oauth", request.BindCreate(authController.OAuthAuth))
```

**Flow:**
1. Frontend initiates OAuth flow with provider (Google/Twitter/LinkedIn) and receives authorization code
2. Frontend sends code and provider name to `/auth/oauth`
3. Backend uses the appropriate provider to exchange code for user info
4. Backend upserts user and returns JWT access and refresh tokens

**Configuration:**
```yaml
auth:
  oauth:
    google:
      client_id: "your-google-client-id.apps.googleusercontent.com"
      client_secret: "your-google-client-secret"
      redirect_url: "http://localhost:8080/auth/google/callback"
    twitter:
      client_id: "your-twitter-client-id"
      client_secret: "your-twitter-client-secret"
      redirect_url: "http://localhost:8080/auth/twitter/callback"
    linkedin:
      client_id: "your-linkedin-client-id"
      client_secret: "your-linkedin-client-secret"
      redirect_url: "http://localhost:8080/auth/linkedin/callback"
```

### Adding a New OAuth Provider

To add support for a new OAuth provider (e.g., Twitter, LinkedIn):

1. **Create a new provider file** (e.g., `twitter_oauth_provider.go`):
```go
package auth

import "context"

type twitterOAuthProvider struct {
	config OAuthConfig
	client *http.Client
}

func NewTwitterOAuthProvider(config OAuthConfig) OAuthProvider {
	return &twitterOAuthProvider{
		config: config,
		client: &http.Client{},
	}
}

func (p *twitterOAuthProvider) ProviderName() string {
	return "twitter"
}

func (p *twitterOAuthProvider) ExchangeCode(ctx context.Context, code string) (*OAuthUserInfo, error) {
	// Implement Twitter OAuth code exchange
	// Return OAuthUserInfo with Provider set to "twitter"
}
```

2. **Update the factory function** in `oauth_provider.go`:
```go
func NewOAuthProvider(provider string, config OAuthConfig) OAuthProvider {
	switch provider {
	case "google":
		return NewGoogleOAuthProvider(config)
	case "twitter":
		return NewTwitterOAuthProvider(config)  // Add new provider
	case "linkedin":
		return NewLinkedInOAuthProvider(config)  // Add new provider
	default:
		return nil
	}
}
```

3. **Update validation** in `entities.go`:
```go
Provider string `json:"provider" binding:"required,oneof=google twitter linkedin"`
```

That's it! The controller, service, and DAO are already generic and will work with any provider.

### Token Refresh

```go
// Route registration (same for both OTP and Google OAuth)
r.POST("/auth/token/refresh", request.BindCreate(authController.RefreshToken))
```

### Protecting Routes with JWT Middleware

```go
// Apply to routes that require authentication
r.Use(auth.GinStdMiddleware(conf.Auth))

// Now these routes require valid JWT token
r.GET("/profile", request.BindGet(profileController.Get))
r.POST("/posts", request.BindCreate(postController.Create))
```

## User Model Requirements

Your user model must implement the `UserModel` interface:

```go
type UserModel interface {
    SetPhone(string) UserModel
    SetSignupInfo(SigupInfo) UserModel
    SetOAuthInfo(OAuthUserInfo) UserModel
    PK() int
    ResourceName() string
}
```

### Example User Model

```go
type User struct {
    pgdb.BaseModel

    // OTP authentication fields
    Phone    string `gorm:"uniqueIndex"`
    DialCode string
    Country  string
    Locale   string

    // OAuth fields (Google, Twitter, LinkedIn, etc.)
    Email      string `gorm:"uniqueIndex"`
    Name       string
    Picture    string
    OAuthID    string `gorm:"uniqueIndex"` // Provider-specific user ID
    OAuthProvider string // "google", "twitter", "linkedin", etc.
}

func (u User) ResourceName() string { return "user" }

func (u User) SetPhone(phone string) auth.UserModel {
    u.Phone = phone
    return u
}

func (u User) SetSignupInfo(info auth.SigupInfo) auth.UserModel {
    u.Phone = info.Phone
    u.DialCode = info.DialCode
    u.Country = info.Country
    u.Locale = info.Locale
    return u
}

func (u User) SetOAuthInfo(info auth.OAuthUserInfo) auth.UserModel {
    u.Email = info.Email
    u.Name = info.Name
    u.Picture = info.Picture
    u.OAuthID = info.ProviderID
    u.OAuthProvider = info.Provider
    if info.Locale != "" {
        u.Locale = info.Locale
    }
    return u
}
```

## API Endpoints

### Send OTP
```
POST /auth/otp/send
Content-Type: application/json

{
  "phone": "+1234567890",
  "dial_code": "+1",
  "country": "US",
  "locale": "en"
}

Response 200:
{
  "retry_after": 60,
  "attempt_left": 2
}
```

### Verify OTP
```
POST /auth/otp/verify
Content-Type: application/json

{
  "phone": "+1234567890",
  "otp": "123456",
  "dial_code": "+1",
  "country": "US",
  "locale": "en"
}

Response 200:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 3600,
  "refresh_expires_in": 2592000
}
```

### OAuth (Google, Twitter, LinkedIn, etc.)
```
POST /auth/oauth
Content-Type: application/json

{
  "code": "4/0AY0e-g7X...",
  "provider": "google",
  "locale": "en"
}

Response 200:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 3600,
  "refresh_expires_in": 2592000
}

Error 400 (Invalid provider):
{
  "error": "Invalid parameter: provider",
  "message": "provider 'facebook' not configured or not supported"
}
```

### Refresh Token
```
POST /auth/token/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGc..."
}

Response 200:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 3600,
  "refresh_expires_in": 2592000
}
```

## Configuration Options

### Main Auth Config
```yaml
auth:
  secret_key: "your-jwt-secret-key"
  access_token_validity_seconds: 3600      # 1 hour
  refresh_token_validity_seconds: 2592000  # 30 days
```

### OTP Config
```yaml
auth:
  otp:
    validity_seconds: 300      # 5 minutes
    max_attempts: 3
    retry_after_seconds: 60
    length: 6
    test_phone: "+1234567890"  # Optional: phone number that always receives OTP "000000"
```

### SMS Provider Config (Twilio)
```yaml
auth:
  twilio:
    account_sid: "your-account-sid"
    auth_token: "your-auth-token"
    from_number: "+1234567890"
```

### Google OAuth Config
```yaml
auth:
  google_oauth:
    client_id: "your-client-id.apps.googleusercontent.com"
    client_secret: "your-client-secret"
    redirect_url: "http://localhost:8080/auth/callback"
```

## Error Handling

All errors follow the `apperrors.AppError` interface and return appropriate HTTP status codes:

- `400 Bad Request`: Invalid parameters (wrong OTP, invalid phone, unverified email)
- `404 Not Found`: User not found
- `500 Internal Server Error`: Server-side errors (database, external API failures)

Example error response:
```json
{
  "error": "Invalid parameter: otp",
  "message": "incorrect otp"
}
```

## Security Notes

1. **JWT Secret**: Use a strong, random secret key in production (at least 32 characters)
2. **HTTPS Only**: Always use HTTPS in production to protect tokens in transit
3. **Token Storage**: Store tokens securely on the client (HttpOnly cookies recommended for web)
4. **Email Verification**: Google OAuth only accepts verified emails (`EmailVerified: true`)
5. **Test Phone**: Remove or protect `test_phone` configuration in production
