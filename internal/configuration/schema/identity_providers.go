package schema

import "time"

// IdentityProvidersConfiguration represents the IdentityProviders 2.0 configuration for Authelia.
type IdentityProvidersConfiguration struct {
	OIDC *OpenIDConnectConfiguration `mapstructure:"oidc"`
}

// OpenIDConnectConfiguration configuration for OpenID Connect.
type OpenIDConnectConfiguration struct {
	// This secret must be 32 bytes long
	HMACSecret       string `mapstructure:"hmac_secret"`
	IssuerPrivateKey string `mapstructure:"issuer_private_key"`

	AccessTokenLifespan       time.Duration `mapstructure:"access_token_lifespan"`
	AuthorizeCodeLifespan     time.Duration `mapstructure:"authorize_code_lifespan"`
	IDTokenLifespan           time.Duration `mapstructure:"id_token_lifespan"`
	RefreshTokenLifespan      time.Duration `mapstructure:"refresh_token_lifespan"`
	EnableClientDebugMessages bool          `mapstructure:"enable_client_debug_messages"`
	MinimumParameterEntropy   int           `mapstructure:"minimum_parameter_entropy"`

	Clients []OpenIDConnectClientConfiguration `mapstructure:"clients"`
}

// OpenIDConnectClientConfiguration configuration for an OpenID Connect client.
type OpenIDConnectClientConfiguration struct {
	ID          string `mapstructure:"id"`
	Description string `mapstructure:"description"`
	Secret      string `mapstructure:"secret"`
	Public      bool   `mapstructure:"public"`

	Policy string `mapstructure:"authorization_policy"`

	Audience      []string `mapstructure:"audience"`
	Scopes        []string `mapstructure:"scopes"`
	RedirectURIs  []string `mapstructure:"redirect_uris"`
	GrantTypes    []string `mapstructure:"grant_types"`
	ResponseTypes []string `mapstructure:"response_types"`
	ResponseModes []string `mapstructure:"response_modes"`

	UserinfoSigningAlgorithm string `mapstructure:"userinfo_signing_algorithm"`
}

// DefaultOpenIDConnectConfiguration contains defaults for OIDC.
var DefaultOpenIDConnectConfiguration = OpenIDConnectConfiguration{
	AccessTokenLifespan:   time.Hour,
	AuthorizeCodeLifespan: time.Minute,
	IDTokenLifespan:       time.Hour,
	RefreshTokenLifespan:  time.Minute * 90,
}

// DefaultOpenIDConnectClientConfiguration contains defaults for OIDC Clients.
var DefaultOpenIDConnectClientConfiguration = OpenIDConnectClientConfiguration{
	Policy:        "two_factor",
	Scopes:        []string{"openid", "groups", "profile", "email"},
	GrantTypes:    []string{"refresh_token", "authorization_code"},
	ResponseTypes: []string{"code"},
	ResponseModes: []string{"form_post", "query", "fragment"},

	UserinfoSigningAlgorithm: "none",
}
