package oidc

import (
	"crypto/rsa"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/herodot"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/internal/authorization"
)

// OpenIDConnectProvider for OpenID Connect.
type OpenIDConnectProvider struct {
	Fosite     fosite.OAuth2Provider
	Store      *OpenIDConnectStore
	KeyManager *KeyManager

	herodot *herodot.JSONWriter
}

// OpenIDConnectStore is Authelia's internal representation of the fosite.Storage interface.
//
//	Currently it is mostly just implementing a decorator pattern other then GetInternalClient.
//	The long term plan is to have these methods interact with the Authelia storage and
//	session providers where applicable.
type OpenIDConnectStore struct {
	clients map[string]*InternalClient
	memory  *storage.MemoryStore
}

// InternalClient represents the client internally.
type InternalClient struct {
	ID          string `json:"id"`
	Description string `json:"-"`
	Secret      []byte `json:"client_secret,omitempty"`
	Public      bool   `json:"public"`

	Policy authorization.Level `json:"-"`

	Audience      []string                  `json:"audience"`
	Scopes        []string                  `json:"scopes"`
	RedirectURIs  []string                  `json:"redirect_uris"`
	GrantTypes    []string                  `json:"grant_types"`
	ResponseTypes []string                  `json:"response_types"`
	ResponseModes []fosite.ResponseModeType `json:"response_modes"`

	UserinfoSigningAlgorithm string `json:"userinfo_signed_response_alg,omitempty"`
}

// KeyManager keeps track of all of the active/inactive rsa keys and provides them to services requiring them.
// It additionally allows us to add keys for the purpose of key rotation in the future.
type KeyManager struct {
	activeKeyID string
	keys        map[string]*rsa.PrivateKey
	keySet      *jose.JSONWebKeySet
	strategy    *RS256JWTStrategy
}

// AutheliaHasher implements the fosite.Hasher interface without an actual hashing algo.
type AutheliaHasher struct{}

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
//
// See https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
type WellKnownConfiguration struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`

	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	RevocationEndpoint    string `json:"revocation_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`

	Algorithms         []string `json:"id_token_signing_alg_values_supported"`
	UserinfoAlgorithms []string `json:"userinfo_signing_alg_values_supported"`

	SubjectTypesSupported  []string `json:"subject_types_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
	ResponseModesSupported []string `json:"response_modes_supported"`
	ScopesSupported        []string `json:"scopes_supported"`
	ClaimsSupported        []string `json:"claims_supported"`

	RequestURIParameterSupported       bool `json:"request_uri_parameter_supported"`
	BackChannelLogoutSupported         bool `json:"backchannel_logout_supported"`
	FrontChannelLogoutSupported        bool `json:"frontchannel_logout_supported"`
	BackChannelLogoutSessionSupported  bool `json:"backchannel_logout_session_supported"`
	FrontChannelLogoutSessionSupported bool `json:"frontchannel_logout_session_supported"`
}

// OpenIDSession holds OIDC Session information.
type OpenIDSession struct {
	*openid.DefaultSession `json:"idToken"`

	Extra    map[string]interface{} `json:"extra"`
	ClientID string
}
