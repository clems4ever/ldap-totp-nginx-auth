package schema

import "time"

// IdentityProvidersConfiguration represents the IdentityProviders 2.0 configuration for Authelia.
type IdentityProvidersConfiguration struct {
	OIDC *OpenIDConnectConfiguration `koanf:"oidc"`
}

// OpenIDConnectConfiguration configuration for OpenID Connect.
type OpenIDConnectConfiguration struct {
	// This secret must be 32 bytes long
	HMACSecret       string `koanf:"hmac_secret"`
	IssuerPrivateKey string `koanf:"issuer_private_key"`

	AccessTokenLifespan   time.Duration `koanf:"access_token_lifespan"`
	AuthorizeCodeLifespan time.Duration `koanf:"authorize_code_lifespan"`
	IDTokenLifespan       time.Duration `koanf:"id_token_lifespan"`
	RefreshTokenLifespan  time.Duration `koanf:"refresh_token_lifespan"`

	EnableClientDebugMessages bool `koanf:"enable_client_debug_messages"`
	MinimumParameterEntropy   int  `koanf:"minimum_parameter_entropy"`

	Clients []OpenIDConnectClientConfiguration `koanf:"clients"`
}

// OpenIDConnectClientConfiguration configuration for an OpenID Connect client.
type OpenIDConnectClientConfiguration struct {
	ID            string   `koanf:"id"`
	Description   string   `koanf:"description"`
	Secret        string   `koanf:"secret"`
	RedirectURIs  []string `koanf:"redirect_uris"`
	Policy        string   `koanf:"authorization_policy"`
	Scopes        []string `koanf:"scopes"`
	GrantTypes    []string `koanf:"grant_types"`
	ResponseTypes []string `koanf:"response_types"`
	ResponseModes []string `koanf:"response_modes"`

	UserinfoSigningAlgorithm string `koanf:"userinfo_signing_algorithm"`
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
