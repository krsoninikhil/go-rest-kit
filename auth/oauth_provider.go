package auth

import "context"

// OAuthProvider defines the interface that all OAuth providers must implement
type OAuthProvider interface {
	// ExchangeCode exchanges an authorization code for user information
	ExchangeCode(ctx context.Context, code string) (*OAuthUserInfo, error)

	// ProviderName returns the name of the provider (e.g., "google", "twitter", "linkedin")
	ProviderName() string
}

// OAuthProviderFactory creates OAuth providers based on configuration
func NewOAuthProvider(provider string, config OAuthConfig) OAuthProvider {
	switch provider {
	case "google":
		return NewGoogleOAuthProvider(config)
	// Add more providers here as they are implemented
	// case "twitter":
	//     return NewTwitterOAuthProvider(config)
	// case "linkedin":
	//     return NewLinkedInOAuthProvider(config)
	default:
		return nil
	}
}
