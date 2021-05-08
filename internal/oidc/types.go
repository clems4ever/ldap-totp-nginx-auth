package oidc

// ConsentGetResponseBody schema of the response body of the consent GET endpoint.
type ConsentGetResponseBody struct {
	ClientID          string     `json:"client_id"`
	ClientDescription string     `json:"client_description"`
	Scopes            []Scope    `json:"scopes"`
	Audience          []Audience `json:"audience"`
}

// Scope represents the scope information.
type Scope struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Audience represents the audience information.
type Audience struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WellKnownConfiguration is the OIDC well known config struct.
type WellKnownConfiguration struct {
	Issuer                             string   `json:"issuer"`
	AuthURL                            string   `json:"authorization_endpoint"`
	TokenURL                           string   `json:"token_endpoint"`
	RevocationEndpoint                 string   `json:"revocation_endpoint"`
	UserinfoEndpoint                   string   `json:"userinfo_endpoint"`
	JWKSURL                            string   `json:"jwks_uri"`
	Algorithms                         []string `json:"id_token_signing_alg_values_supported"`
	ResponseTypesSupported             []string `json:"response_types_supported"`
	ScopesSupported                    []string `json:"scopes_supported"`
	ClaimsSupported                    []string `json:"claims_supported"`
	BackChannelLogoutSupported         bool     `json:"backchannel_logout_supported"`
	BackChannelLogoutSessionSupported  bool     `json:"backchannel_logout_session_supported"`
	FrontChannelLogoutSupported        bool     `json:"frontchannel_logout_supported"`
	FrontChannelLogoutSessionSupported bool     `json:"frontchannel_logout_session_supported"`
}