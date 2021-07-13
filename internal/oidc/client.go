package oidc

import (
	"github.com/ory/fosite"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/session"
)

// NewClient creates a new InternalClient.
func NewClient(config schema.OpenIDConnectClientConfiguration) (client *InternalClient) {
	client = &InternalClient{
		ID:          config.ID,
		Description: config.Description,
		Secret:      []byte(config.Secret),
		Public:      config.Public,

		Policy: authorization.PolicyToLevel(config.Policy),

		Audience:      config.Audience,
		Scopes:        config.Scopes,
		RedirectURIs:  config.RedirectURIs,
		GrantTypes:    config.GrantTypes,
		ResponseTypes: config.ResponseTypes,
		ResponseModes: []fosite.ResponseModeType{fosite.ResponseModeDefault},

		UserinfoSigningAlgorithm: config.UserinfoSigningAlgorithm,
	}

	for _, mode := range config.ResponseModes {
		client.ResponseModes = append(client.ResponseModes, fosite.ResponseModeType(mode))
	}

	return client
}

// IsAuthenticationLevelSufficient returns if the provided authentication.Level is sufficient for the client of the AutheliaClient.
func (c InternalClient) IsAuthenticationLevelSufficient(level authentication.Level) bool {
	return authorization.IsAuthLevelSufficient(level, c.Policy)
}

// GetID returns the ID.
func (c InternalClient) GetID() string {
	return c.ID
}

// GetConsentResponseBody returns the proper consent response body for this session.OIDCWorkflowSession.
func (c InternalClient) GetConsentResponseBody(session *session.OIDCWorkflowSession) ConsentGetResponseBody {
	body := ConsentGetResponseBody{
		ClientID:          c.ID,
		ClientDescription: c.Description,
	}

	if session != nil {
		body.Scopes = scopeNamesToScopes(session.RequestedScopes)
		body.Audience = audienceNamesToAudience(session.RequestedAudience)
	}

	return body
}

// GetHashedSecret returns the Secret.
func (c InternalClient) GetHashedSecret() []byte {
	return c.Secret
}

// GetRedirectURIs returns the RedirectURIs.
func (c InternalClient) GetRedirectURIs() []string {
	return c.RedirectURIs
}

// GetGrantTypes returns the GrantTypes.
func (c InternalClient) GetGrantTypes() fosite.Arguments {
	if len(c.GrantTypes) == 0 {
		return fosite.Arguments{"authorization_code"}
	}

	return c.GrantTypes
}

// GetResponseTypes returns the ResponseTypes.
func (c InternalClient) GetResponseTypes() fosite.Arguments {
	if len(c.ResponseTypes) == 0 {
		return fosite.Arguments{"code"}
	}

	return c.ResponseTypes
}

// GetScopes returns the Scopes.
func (c InternalClient) GetScopes() fosite.Arguments {
	return c.Scopes
}

// IsPublic returns the value of the Public property.
func (c InternalClient) IsPublic() bool {
	return c.Public
}

// GetAudience returns the Audience.
func (c InternalClient) GetAudience() fosite.Arguments {
	return c.Audience
}

// GetResponseModes returns the valid response modes for this client.
//
// Implements the fosite.ResponseModeClient.
func (c InternalClient) GetResponseModes() []fosite.ResponseModeType {
	return c.ResponseModes
}
