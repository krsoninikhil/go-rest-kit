# OAuth Integration Example

This example demonstrates how to integrate OAuth authentication with Google, and shows how easy it is to add more providers.

## Complete Example

```go
package main

import (
    "context"
    "log"

    "github.com/gin-gonic/gin"
    "github.com/krsoninikhil/go-rest-kit/auth"
    "github.com/krsoninikhil/go-rest-kit/cache"
    "github.com/krsoninikhil/go-rest-kit/config"
    "github.com/krsoninikhil/go-rest-kit/integrations/twilio"
    "github.com/krsoninikhil/go-rest-kit/pgdb"
    "github.com/krsoninikhil/go-rest-kit/request"
)

// User model that supports both OTP and OAuth authentication
type User struct {
    pgdb.BaseModel

    // OTP authentication fields
    Phone    string `gorm:"uniqueIndex"`
    DialCode string
    Country  string

    // OAuth fields (works with any provider: Google, Twitter, LinkedIn, etc.)
    Email         string `gorm:"uniqueIndex"`
    Name          string
    Picture       string
    OAuthID       string `gorm:"uniqueIndex"` // Provider-specific user ID
    OAuthProvider string                      // "google", "twitter", "linkedin", etc.

    Locale string
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

// This single method supports ALL OAuth providers!
func (u User) SetOAuthInfo(info auth.OAuthUserInfo) auth.UserModel {
    u.Email = info.Email
    u.Name = info.Name
    u.Picture = info.Picture
    u.OAuthID = info.ProviderID
    u.OAuthProvider = info.Provider // Automatically set based on provider used
    if info.Locale != "" {
        u.Locale = info.Locale
    }
    return u
}

// Configuration
type Config struct {
    DB   pgdb.Config
    Auth auth.Config
    Env  string
}

func (c *Config) EnvPath() string    { return "./.env" }
func (c *Config) SourcePath() string { return fmt.Sprintf("./config/%s.yml", c.Env) }
func (c *Config) SetEnv(env string)  { c.Env = env }

func main() {
    var (
        ctx  = context.Background()
        conf Config
    )
    config.Load(ctx, &conf)

    // Database connection
    db := pgdb.NewPGConnection(ctx, conf.DB)

    // Auth service setup
    cache := cache.NewInMemory()
    userDao := auth.NewUserDao[User](db)
    authSvc := auth.NewService(conf.Auth, userDao)

    // OTP service
    smsProvider := twilio.NewClient(conf.Auth.Twilio)
    otpSvc := auth.NewOTPSvc(conf.Auth.OTP, smsProvider, cache)

    // OAuth providers setup
    var oauthProviders []auth.OAuthProvider

    // Add Google if configured
    if googleConfig, exists := conf.Auth.OAuth["google"]; exists {
        googleProvider := auth.NewGoogleOAuthProvider(googleConfig)
        oauthProviders = append(oauthProviders, googleProvider)
        log.Printf("Google OAuth provider enabled")
    }

    // Add Twitter if configured (when implemented)
    // if twitterConfig, exists := conf.Auth.OAuth["twitter"]; exists {
    //     twitterProvider := auth.NewTwitterOAuthProvider(twitterConfig)
    //     oauthProviders = append(oauthProviders, twitterProvider)
    //     log.Printf("Twitter OAuth provider enabled")
    // }

    // Add LinkedIn if configured (when implemented)
    // if linkedinConfig, exists := conf.Auth.OAuth["linkedin"]; exists {
    //     linkedinProvider := auth.NewLinkedInOAuthProvider(linkedinConfig)
    //     oauthProviders = append(oauthProviders, linkedinProvider)
    //     log.Printf("LinkedIn OAuth provider enabled")
    // }

    // Create auth controller with all providers
    authController := auth.NewController(authSvc, otpSvc, cache).
        WithOAuthProviders(oauthProviders...)

    // Setup routes
    r := gin.Default()

    // OTP authentication routes
    r.POST("/auth/otp/send", request.BindCreate(authController.SendOTP))
    r.POST("/auth/otp/verify", request.BindCreate(authController.VerifyOTP))

    // Single OAuth route handles ALL providers (Google, Twitter, LinkedIn, etc.)
    r.POST("/auth/oauth", request.BindCreate(authController.OAuthAuth))

    // Token refresh (works for both OTP and OAuth users)
    r.POST("/auth/token/refresh", request.BindCreate(authController.RefreshToken))

    // Protected routes
    r.Use(auth.GinStdMiddleware(conf.Auth))
    r.GET("/profile", request.BindGet(getProfile))

    // Start server
    if err := r.Run(":8080"); err != nil {
        log.Fatal("Could not start server:", err)
    }
}

func getProfile(c *gin.Context) (*User, error) {
    // Your profile logic here
    return nil, nil
}
```

## Configuration File

`config/development.yml`:
```yaml
db:
  host: localhost
  port: 5432
  database: myapp
  username: postgres
  password: ""

auth:
  secret_key: "your-super-secret-jwt-key-here"
  access_token_validity_seconds: 3600
  refresh_token_validity_seconds: 2592000

  # OTP configuration
  otp:
    validity_seconds: 300
    max_attempts: 3
    retry_after_seconds: 60
    length: 6
    test_phone: "+1234567890"

  # SMS provider (for OTP)
  twilio:
    account_sid: "your-twilio-account-sid"
    auth_token: "your-twilio-auth-token"
    from_number: "+1234567890"

  # OAuth providers
  oauth:
    google:
      client_id: "your-google-client-id.apps.googleusercontent.com"
      client_secret: "your-google-client-secret"
      redirect_url: "http://localhost:8080/auth/google/callback"

    # Add more providers as you implement them
    # twitter:
    #   client_id: "your-twitter-api-key"
    #   client_secret: "your-twitter-api-secret"
    #   redirect_url: "http://localhost:8080/auth/twitter/callback"

    # linkedin:
    #   client_id: "your-linkedin-client-id"
    #   client_secret: "your-linkedin-client-secret"
    #   redirect_url: "http://localhost:8080/auth/linkedin/callback"
```

## Frontend Usage

### Google OAuth
```javascript
// 1. Get authorization code from Google
const googleAuthCode = await getGoogleAuthCode();

// 2. Exchange code for JWT tokens
const response = await fetch('http://localhost:8080/auth/oauth', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    code: googleAuthCode,
    provider: 'google',
    locale: 'en'
  })
});

const { access_token, refresh_token } = await response.json();
```

### Twitter OAuth (when implemented)
```javascript
const twitterAuthCode = await getTwitterAuthCode();

const response = await fetch('http://localhost:8080/auth/oauth', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    code: twitterAuthCode,
    provider: 'twitter',  // Just change the provider!
    locale: 'en'
  })
});
```

## Benefits of This Design

1. **Single Endpoint**: One `/auth/oauth` endpoint handles all providers
2. **Easy to Extend**: Adding a new provider requires:
   - Implementing the `OAuthProvider` interface
   - Adding to the factory function
   - Updating configuration
   - No changes to controller, service, or DAO!
3. **Consistent User Experience**: All OAuth users are handled the same way
4. **Flexible Configuration**: Enable/disable providers via config without code changes
5. **Type Safety**: Generic types ensure compile-time safety across all providers

## Adding Twitter OAuth Support

Here's a template for adding Twitter:

```go
// twitter_oauth_provider.go
package auth

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"

    "github.com/krsoninikhil/go-rest-kit/apperrors"
    "github.com/pkg/errors"
)

const (
    twitterTokenURL    = "https://api.twitter.com/2/oauth2/token"
    twitterUserInfoURL = "https://api.twitter.com/2/users/me"
    twitterProviderName = "twitter"
)

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
    return twitterProviderName
}

func (p *twitterOAuthProvider) ExchangeCode(ctx context.Context, code string) (*OAuthUserInfo, error) {
    // 1. Exchange code for access token
    accessToken, err := p.exchangeCodeForToken(ctx, code)
    if err != nil {
        return nil, err
    }

    // 2. Get user info
    userInfo, err := p.getUserInfo(ctx, accessToken)
    if err != nil {
        return nil, err
    }

    return userInfo, nil
}

// Implement exchangeCodeForToken and getUserInfo similar to Google provider
// ...
```

Then update `oauth_provider.go`:
```go
func NewOAuthProvider(provider string, config OAuthConfig) OAuthProvider {
    switch provider {
    case "google":
        return NewGoogleOAuthProvider(config)
    case "twitter":
        return NewTwitterOAuthProvider(config)  // Add this line
    default:
        return nil
    }
}
```

And update validation in `entities.go`:
```go
Provider string `json:"provider" binding:"required,oneof=google twitter"`
```

That's it! Your existing controller, service, and DAO will automatically work with Twitter.
