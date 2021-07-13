package validator

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldRaiseErrorWhenInvalidOIDCServerConfiguration(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "abc",
			IssuerPrivateKey: "",
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 2)

	assert.EqualError(t, validator.Errors()[0], "OIDC Server issuer private key must be provided")
	assert.EqualError(t, validator.Errors()[1], "OIDC Server has no clients defined")
}

func TestShouldRaiseErrorWhenOIDCServerIssuerPrivateKeyPathInvalid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "OIDC Server has no clients defined")
}

func TestShouldRaiseErrorWhenOIDCServerClientBadValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "",
					Secret: "",
					Policy: "",
					RedirectURIs: []string{
						"tcp://google.com",
					},
				},
				{
					ID:     "a-client",
					Secret: "a-secret",
					Policy: "a-policy",
					RedirectURIs: []string{
						"https://google.com",
					},
				},
				{
					ID:     "a-client",
					Secret: "a-secret",
					Policy: "a-policy",
					RedirectURIs: []string{
						"https://google.com",
					},
				},
				{
					ID:     "client-check-uri-parse",
					Secret: "a-secret",
					Policy: twoFactorPolicy,
					RedirectURIs: []string{
						"http://abc@%two",
					},
				},
				{
					ID:     "client-check-uri-abs",
					Secret: "a-secret",
					Policy: twoFactorPolicy,
					RedirectURIs: []string{
						"google.com",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 8)

	assert.Equal(t, schema.DefaultOpenIDConnectClientConfiguration.Policy, config.OIDC.Clients[0].Policy)
	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtOIDCServerClientInvalidSecret, ""))
	assert.EqualError(t, validator.Errors()[1], fmt.Sprintf(errFmtOIDCServerClientRedirectURI, "", "tcp://google.com", "tcp"))
	assert.EqualError(t, validator.Errors()[2], fmt.Sprintf(errFmtOIDCServerClientInvalidPolicy, "a-client", "a-policy"))
	assert.EqualError(t, validator.Errors()[3], fmt.Sprintf(errFmtOIDCServerClientInvalidPolicy, "a-client", "a-policy"))
	assert.EqualError(t, validator.Errors()[4], fmt.Sprintf(errFmtOIDCServerClientRedirectURICantBeParsed, "client-check-uri-parse", "http://abc@%two", errors.New("parse \"http://abc@%two\": invalid URL escape \"%tw\"")))
	assert.EqualError(t, validator.Errors()[5], fmt.Sprintf(errFmtOIDCClientRedirectURIAbsolute, "client-check-uri-abs", "google.com"))
	assert.EqualError(t, validator.Errors()[6], "OIDC Server has one or more clients with an empty ID")
	assert.EqualError(t, validator.Errors()[7], "OIDC Server has clients with duplicate ID's")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadScopes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "good_id",
					Secret: "good_secret",
					Policy: "two_factor",
					Scopes: []string{"openid", "bad_scope"},
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "OIDC client with ID 'good_id' has an invalid scope "+
		"'bad_scope', must be one of: 'openid', 'email', 'profile', 'groups', 'offline_access'")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadGrantTypes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:         "good_id",
					Secret:     "good_secret",
					Policy:     "two_factor",
					GrantTypes: []string{"bad_grant_type"},
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "OIDC client with ID 'good_id' has an invalid grant type "+
		"'bad_grant_type', must be one of: 'implicit', 'refresh_token', 'authorization_code', "+
		"'password', 'client_credentials'")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadResponseModes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:            "good_id",
					Secret:        "good_secret",
					Policy:        "two_factor",
					ResponseModes: []string{"bad_responsemode"},
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "OIDC client with ID 'good_id' has an invalid response mode "+
		"'bad_responsemode', must be one of: 'form_post', 'query', 'fragment'")
}

func TestShouldRaiseErrorWhenOIDCClientConfiguredWithBadUserinfoAlg(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "key-material",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:                       "good_id",
					Secret:                   "good_secret",
					Policy:                   "two_factor",
					UserinfoSigningAlgorithm: "rs256",
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "OIDC client with ID 'good_id' has an invalid userinfo "+
		"signing algorithm 'rs256', must be one of: 'none, RS256'")
}

func TestValidateIdentityProvidersShouldRaiseWarningOnSecurityIssue(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:              "abc",
			IssuerPrivateKey:        "abc",
			MinimumParameterEntropy: 1,
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "good_id",
					Secret: "good_secret",
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://google.com/callback",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Errors(), 0)
	require.Len(t, validator.Warnings(), 1)

	assert.EqualError(t, validator.Warnings()[0], "SECURITY ISSUE: OIDC minimum parameter entropy is configured to an unsafe value, it should be above 8 but it's configured to 1.")
}

func TestValidateIdentityProvidersShouldRaiseErrorsOnInvalidPublicClients(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "hmac1",
			IssuerPrivateKey: "key2",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "client-with-invalid-secret",
					Secret: "a-secret",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://localhost",
					},
				},
				{
					ID:     "client-with-invalid-redirect-path",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://localhost/callback",
					},
				},
				{
					ID:     "client-with-invalid-redirect-queryparams",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://localhost?api=x",
					},
				},
				{
					ID:     "client-with-invalid-redirect-scheme",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"ftp://localhost",
					},
				},
				{
					ID:     "client-with-invalid-redirect-noscheme",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						localhost,
					},
				},
				{
					ID:     "client-with-invalid-redirect-host",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://google.com",
					},
				},
				{
					ID:     "client-with-invalid-redirect-fragment",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://localhost#fragment",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	require.Len(t, validator.Errors(), 7)
	assert.Len(t, validator.Warnings(), 0)

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtOIDCClientPublicInvalidSecret, "client-with-invalid-secret"))
	assert.EqualError(t, validator.Errors()[1], fmt.Sprintf(errFmtOIDCClientPublicRedirectURIPath, "client-with-invalid-redirect-path", "https://localhost/callback"))
	assert.EqualError(t, validator.Errors()[2], fmt.Sprintf(errFmtOIDCClientPublicRedirectURIQuery, "client-with-invalid-redirect-queryparams", "https://localhost?api=x"))
	assert.EqualError(t, validator.Errors()[3], fmt.Sprintf(errFmtOIDCServerClientRedirectURI, "client-with-invalid-redirect-scheme", "ftp://localhost", "ftp"))
	assert.EqualError(t, validator.Errors()[4], fmt.Sprintf(errFmtOIDCClientRedirectURIAbsolute, "client-with-invalid-redirect-noscheme", "localhost"))
	assert.EqualError(t, validator.Errors()[5], fmt.Sprintf(errFmtOIDCClientPublicRedirectURIHost, "client-with-invalid-redirect-host", "https://google.com"))
	assert.EqualError(t, validator.Errors()[6], fmt.Sprintf(errFmtOIDCClientPublicRedirectURIFragment, "client-with-invalid-redirect-fragment", "https://localhost#fragment"))
}

func TestValidateIdentityProvidersShouldNotRaiseErrorsOnValidPublicClients(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "hmac1",
			IssuerPrivateKey: "key2",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "installed-app-client",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						oauth2InstalledApp,
					},
				},
				{
					ID:     "client-with-https-scheme",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"https://localhost:9000",
					},
				},
				{
					ID:     "client-with-loopback",
					Public: true,
					Policy: "two_factor",
					RedirectURIs: []string{
						"http://127.0.0.1",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Len(t, validator.Warnings(), 0)
}

func TestValidateIdentityProvidersShouldSetDefaultValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := &schema.IdentityProvidersConfiguration{
		OIDC: &schema.OpenIDConnectConfiguration{
			HMACSecret:       "rLABDrx87et5KvRHVUgTm3pezWWd8LMN",
			IssuerPrivateKey: "../../../README.md",
			Clients: []schema.OpenIDConnectClientConfiguration{
				{
					ID:     "a-client",
					Secret: "a-client-secret",
					RedirectURIs: []string{
						"https://google.com",
					},
				},
				{
					ID:                       "b-client",
					Description:              "Normal Description",
					Secret:                   "b-client-secret",
					Policy:                   oneFactorPolicy,
					UserinfoSigningAlgorithm: "RS256",
					RedirectURIs: []string{
						"https://google.com",
					},
					Scopes: []string{
						"groups",
					},
					GrantTypes: []string{
						"refresh_token",
					},
					ResponseTypes: []string{
						"token",
						"code",
					},
					ResponseModes: []string{
						"form_post",
						"fragment",
					},
				},
			},
		},
	}

	ValidateIdentityProviders(config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	// Assert Clients[0] Policy is set to the default, and the default doesn't override Clients[1]'s Policy.
	assert.Equal(t, config.OIDC.Clients[0].Policy, twoFactorPolicy)
	assert.Equal(t, config.OIDC.Clients[1].Policy, oneFactorPolicy)

	assert.Equal(t, config.OIDC.Clients[0].UserinfoSigningAlgorithm, "none")
	assert.Equal(t, config.OIDC.Clients[1].UserinfoSigningAlgorithm, "RS256")

	// Assert Clients[0] Description is set to the Clients[0] ID, and Clients[1]'s Description is not overridden.
	assert.Equal(t, config.OIDC.Clients[0].ID, config.OIDC.Clients[0].Description)
	assert.Equal(t, "Normal Description", config.OIDC.Clients[1].Description)

	// Assert Clients[0] ends up configured with the default Scopes.
	require.Len(t, config.OIDC.Clients[0].Scopes, 4)
	assert.Equal(t, "openid", config.OIDC.Clients[0].Scopes[0])
	assert.Equal(t, "groups", config.OIDC.Clients[0].Scopes[1])
	assert.Equal(t, "profile", config.OIDC.Clients[0].Scopes[2])
	assert.Equal(t, "email", config.OIDC.Clients[0].Scopes[3])

	// Assert Clients[1] ends up configured with the configured Scopes and the openid Scope.
	require.Len(t, config.OIDC.Clients[1].Scopes, 2)
	assert.Equal(t, "groups", config.OIDC.Clients[1].Scopes[0])
	assert.Equal(t, "openid", config.OIDC.Clients[1].Scopes[1])

	// Assert Clients[0] ends up configured with the default GrantTypes.
	require.Len(t, config.OIDC.Clients[0].GrantTypes, 2)
	assert.Equal(t, "refresh_token", config.OIDC.Clients[0].GrantTypes[0])
	assert.Equal(t, "authorization_code", config.OIDC.Clients[0].GrantTypes[1])

	// Assert Clients[1] ends up configured with only the configured GrantTypes.
	require.Len(t, config.OIDC.Clients[1].GrantTypes, 1)
	assert.Equal(t, "refresh_token", config.OIDC.Clients[1].GrantTypes[0])

	// Assert Clients[0] ends up configured with the default ResponseTypes.
	require.Len(t, config.OIDC.Clients[0].ResponseTypes, 1)
	assert.Equal(t, "code", config.OIDC.Clients[0].ResponseTypes[0])

	// Assert Clients[1] ends up configured only with the configured ResponseTypes.
	require.Len(t, config.OIDC.Clients[1].ResponseTypes, 2)
	assert.Equal(t, "token", config.OIDC.Clients[1].ResponseTypes[0])
	assert.Equal(t, "code", config.OIDC.Clients[1].ResponseTypes[1])

	// Assert Clients[0] ends up configured with the default ResponseModes.
	require.Len(t, config.OIDC.Clients[0].ResponseModes, 3)
	assert.Equal(t, "form_post", config.OIDC.Clients[0].ResponseModes[0])
	assert.Equal(t, "query", config.OIDC.Clients[0].ResponseModes[1])
	assert.Equal(t, "fragment", config.OIDC.Clients[0].ResponseModes[2])

	// Assert Clients[1] ends up configured only with the configured ResponseModes.
	require.Len(t, config.OIDC.Clients[1].ResponseModes, 2)
	assert.Equal(t, "form_post", config.OIDC.Clients[1].ResponseModes[0])
	assert.Equal(t, "fragment", config.OIDC.Clients[1].ResponseModes[1])

	assert.Equal(t, false, config.OIDC.EnableClientDebugMessages)
	assert.Equal(t, time.Hour, config.OIDC.AccessTokenLifespan)
	assert.Equal(t, time.Minute, config.OIDC.AuthorizeCodeLifespan)
	assert.Equal(t, time.Hour, config.OIDC.IDTokenLifespan)
	assert.Equal(t, time.Minute*90, config.OIDC.RefreshTokenLifespan)
}
