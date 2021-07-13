package validator

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateIdentityProviders validates and update IdentityProviders configuration.
func ValidateIdentityProviders(configuration *schema.IdentityProvidersConfiguration, validator *schema.StructValidator) {
	validateOIDC(configuration.OIDC, validator)
}

func validateOIDC(configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if configuration != nil {
		if configuration.IssuerPrivateKey == "" {
			validator.Push(fmt.Errorf("OIDC Server issuer private key must be provided"))
		}

		if configuration.AccessTokenLifespan == time.Duration(0) {
			configuration.AccessTokenLifespan = schema.DefaultOpenIDConnectConfiguration.AccessTokenLifespan
		}

		if configuration.AuthorizeCodeLifespan == time.Duration(0) {
			configuration.AuthorizeCodeLifespan = schema.DefaultOpenIDConnectConfiguration.AuthorizeCodeLifespan
		}

		if configuration.IDTokenLifespan == time.Duration(0) {
			configuration.IDTokenLifespan = schema.DefaultOpenIDConnectConfiguration.IDTokenLifespan
		}

		if configuration.RefreshTokenLifespan == time.Duration(0) {
			configuration.RefreshTokenLifespan = schema.DefaultOpenIDConnectConfiguration.RefreshTokenLifespan
		}

		if configuration.MinimumParameterEntropy != 0 && configuration.MinimumParameterEntropy < 8 {
			validator.PushWarning(fmt.Errorf(errFmtOIDCServerInsecureParameterEntropy, configuration.MinimumParameterEntropy))
		}

		validateOIDCClients(configuration, validator)

		if len(configuration.Clients) == 0 {
			validator.Push(fmt.Errorf("OIDC Server has no clients defined"))
		}
	}
}

func validateOIDCClients(configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	invalidID, duplicateIDs := false, false

	var ids []string

	for c, client := range configuration.Clients {
		if client.ID == "" {
			invalidID = true
		} else {
			if client.Description == "" {
				configuration.Clients[c].Description = client.ID
			}

			if utils.IsStringInSliceFold(client.ID, ids) {
				duplicateIDs = true
			}
			ids = append(ids, client.ID)
		}

		if client.Public {
			if client.Secret != "" {
				validator.Push(fmt.Errorf(errFmtOIDCClientPublicInvalidSecret, client.ID))
			}
		} else {
			if client.Secret == "" {
				validator.Push(fmt.Errorf(errFmtOIDCServerClientInvalidSecret, client.ID))
			}
		}

		if client.Policy == "" {
			configuration.Clients[c].Policy = schema.DefaultOpenIDConnectClientConfiguration.Policy
		} else if client.Policy != oneFactorPolicy && client.Policy != twoFactorPolicy {
			validator.Push(fmt.Errorf(errFmtOIDCServerClientInvalidPolicy, client.ID, client.Policy))
		}

		validateOIDCClientScopes(c, configuration, validator)
		validateOIDCClientGrantTypes(c, configuration, validator)
		validateOIDCClientResponseTypes(c, configuration, validator)
		validateOIDCClientResponseModes(c, configuration, validator)
		validateOIDDClientUserinfoAlgorithm(c, configuration, validator)

		validateOIDCClientRedirectURIs(client, validator)
	}

	if invalidID {
		validator.Push(fmt.Errorf("OIDC Server has one or more clients with an empty ID"))
	}

	if duplicateIDs {
		validator.Push(fmt.Errorf("OIDC Server has clients with duplicate ID's"))
	}
}

func validateOIDCClientScopes(c int, configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if len(configuration.Clients[c].Scopes) == 0 {
		configuration.Clients[c].Scopes = schema.DefaultOpenIDConnectClientConfiguration.Scopes
		return
	}

	if !utils.IsStringInSlice("openid", configuration.Clients[c].Scopes) {
		configuration.Clients[c].Scopes = append(configuration.Clients[c].Scopes, "openid")
	}

	for _, scope := range configuration.Clients[c].Scopes {
		if !utils.IsStringInSlice(scope, validOIDCScopes) {
			validator.Push(fmt.Errorf(
				errFmtOIDCServerClientInvalidScope,
				configuration.Clients[c].ID, scope, strings.Join(validOIDCScopes, "', '")))
		}
	}
}

func validateOIDCClientGrantTypes(c int, configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if len(configuration.Clients[c].GrantTypes) == 0 {
		configuration.Clients[c].GrantTypes = schema.DefaultOpenIDConnectClientConfiguration.GrantTypes
		return
	}

	for _, grantType := range configuration.Clients[c].GrantTypes {
		if !utils.IsStringInSlice(grantType, validOIDCGrantTypes) {
			validator.Push(fmt.Errorf(
				errFmtOIDCServerClientInvalidGrantType,
				configuration.Clients[c].ID, grantType, strings.Join(validOIDCGrantTypes, "', '")))
		}
	}
}

func validateOIDCClientResponseTypes(c int, configuration *schema.OpenIDConnectConfiguration, _ *schema.StructValidator) {
	if len(configuration.Clients[c].ResponseTypes) == 0 {
		configuration.Clients[c].ResponseTypes = schema.DefaultOpenIDConnectClientConfiguration.ResponseTypes
		return
	}
}

func validateOIDCClientResponseModes(c int, configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if len(configuration.Clients[c].ResponseModes) == 0 {
		configuration.Clients[c].ResponseModes = schema.DefaultOpenIDConnectClientConfiguration.ResponseModes
		return
	}

	for _, responseMode := range configuration.Clients[c].ResponseModes {
		if !utils.IsStringInSlice(responseMode, validOIDCResponseModes) {
			validator.Push(fmt.Errorf(
				errFmtOIDCServerClientInvalidResponseMode,
				configuration.Clients[c].ID, responseMode, strings.Join(validOIDCResponseModes, "', '")))
		}
	}
}

func validateOIDDClientUserinfoAlgorithm(c int, configuration *schema.OpenIDConnectConfiguration, validator *schema.StructValidator) {
	if configuration.Clients[c].UserinfoSigningAlgorithm == "" {
		configuration.Clients[c].UserinfoSigningAlgorithm = schema.DefaultOpenIDConnectClientConfiguration.UserinfoSigningAlgorithm
	} else if !utils.IsStringInSlice(configuration.Clients[c].UserinfoSigningAlgorithm, validOIDCUserinfoAlgorithms) {
		validator.Push(fmt.Errorf(errFmtOIDCServerClientInvalidUserinfoAlgorithm,
			configuration.Clients[c].ID, configuration.Clients[c].UserinfoSigningAlgorithm, strings.Join(validOIDCUserinfoAlgorithms, ", ")))
	}
}

func validateOIDCClientRedirectURIs(client schema.OpenIDConnectClientConfiguration, validator *schema.StructValidator) {
	for _, redirectURI := range client.RedirectURIs {
		parsedURL, err := url.Parse(redirectURI)
		if err != nil {
			validator.Push(fmt.Errorf(errFmtOIDCServerClientRedirectURICantBeParsed, client.ID, redirectURI, err))
			continue
		}

		if client.Public {
			validateOIDCPublicClientRedirectURI(client.ID, redirectURI, parsedURL, validator)
		} else {
			validateOIDCClientRedirectURI(client.ID, redirectURI, parsedURL, validator)
		}
	}
}

func validateOIDCPublicClientRedirectURI(id, redirectURI string, parsedURL *url.URL, validator *schema.StructValidator) {
	if redirectURI == oauth2InstalledApp {
		return
	}

	host := strings.SplitN(parsedURL.Host, ":", 2)
	if host[0] != "" && host[0] != localhost && host[0] != loopback {
		validator.Push(fmt.Errorf(errFmtOIDCClientPublicRedirectURIHost, id, redirectURI))
	}

	if host[0] != "" && parsedURL.Path != "" || parsedURL.RawPath != "" {
		validator.Push(fmt.Errorf(errFmtOIDCClientPublicRedirectURIPath, id, redirectURI))
	}

	if parsedURL.RawQuery != "" {
		validator.Push(fmt.Errorf(errFmtOIDCClientPublicRedirectURIQuery, id, redirectURI))
	}

	if parsedURL.Fragment != "" || parsedURL.RawFragment != "" {
		validator.Push(fmt.Errorf(errFmtOIDCClientPublicRedirectURIFragment, id, redirectURI))
	}

	if !parsedURL.IsAbs() {
		validator.Push(fmt.Errorf(errFmtOIDCClientRedirectURIAbsolute, id, redirectURI))
		return
	}

	if parsedURL.Scheme != schemeHTTPS && parsedURL.Scheme != schemeHTTP {
		validator.Push(fmt.Errorf(errFmtOIDCServerClientRedirectURI, id, redirectURI, parsedURL.Scheme))
	}
}

func validateOIDCClientRedirectURI(id, redirectURI string, parsedURL *url.URL, validator *schema.StructValidator) {
	if !parsedURL.IsAbs() {
		validator.Push(fmt.Errorf(errFmtOIDCClientRedirectURIAbsolute, id, redirectURI))
		return
	}

	if parsedURL.Scheme != schemeHTTPS && parsedURL.Scheme != schemeHTTP {
		validator.Push(fmt.Errorf(errFmtOIDCServerClientRedirectURI, id, redirectURI, parsedURL.Scheme))
	}
}
