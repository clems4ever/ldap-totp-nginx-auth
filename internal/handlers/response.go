package handlers

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/utils"
)

// handleOIDCWorkflowResponse handle the redirection upon authentication in the OIDC workflow.
func handleOIDCWorkflowResponse(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	if !authorization.IsAuthLevelSufficient(userSession.AuthenticationLevel, userSession.OIDCWorkflowSession.RequiredAuthorizationLevel) {
		ctx.Logger.Warn("OIDC requires 2FA, cannot be redirected yet")
		ctx.ReplyOK()

		return
	}

	uri, err := ctx.ForwardedProtoHost()
	if err != nil {
		ctx.Logger.Errorf("%v", err)
		handleAuthenticationUnauthorized(ctx, fmt.Errorf("Unable to get forward facing URI"), authenticationFailedMessage)

		return
	}

	if isConsentMissing(
		userSession.OIDCWorkflowSession,
		userSession.OIDCWorkflowSession.RequestedScopes,
		userSession.OIDCWorkflowSession.RequestedAudience) {
		err := ctx.SetJSONBody(redirectResponse{Redirect: fmt.Sprintf("%s/consent", uri)})

		if err != nil {
			ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
		}
	} else {
		err := ctx.SetJSONBody(redirectResponse{Redirect: userSession.OIDCWorkflowSession.AuthURI})
		if err != nil {
			ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
		}
	}
}

// Handle1FAResponse handle the redirection upon 1FA authentication.
func Handle1FAResponse(ctx *middlewares.AutheliaCtx, targetURI, requestMethod string, username string, groups []string) {
	if targetURI == "" {
		if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && ctx.Configuration.DefaultRedirectionURL != "" {
			err := ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL})
			if err != nil {
				ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
			}
		} else {
			ctx.ReplyOK()
		}

		return
	}

	targetURL, err := url.ParseRequestURI(targetURI)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to parse target URL %s: %s", targetURI, err), authenticationFailedMessage)
		return
	}

	requiredLevel := ctx.Providers.Authorizer.GetRequiredLevel(
		authorization.Subject{
			Username: username,
			Groups:   groups,
			IP:       ctx.RemoteIP(),
		},
		authorization.NewObject(targetURL, requestMethod))

	ctx.Logger.Debugf("Required level for the URL %s is %d", targetURI, requiredLevel)

	if requiredLevel == authorization.TwoFactor {
		ctx.Logger.Warnf("%s requires 2FA, cannot be redirected yet", targetURI)
		ctx.ReplyOK()

		return
	}

	safeRedirection := utils.IsRedirectionSafe(*targetURL, ctx.Configuration.Session.Domain)

	if !safeRedirection {
		if !ctx.Providers.Authorizer.IsSecondFactorEnabled() && ctx.Configuration.DefaultRedirectionURL != "" {
			err := ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL})
			if err != nil {
				ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
			}
		} else {
			ctx.ReplyOK()
		}

		return
	}

	ctx.Logger.Debugf("Redirection URL %s is safe", targetURI)
	err = ctx.SetJSONBody(redirectResponse{Redirect: targetURI})

	if err != nil {
		ctx.Logger.Errorf("Unable to set redirection URL in body: %s", err)
	}
}

// Handle2FAResponse handle the redirection upon 2FA authentication.
func Handle2FAResponse(ctx *middlewares.AutheliaCtx, targetURI string) {
	if targetURI == "" {
		if ctx.Configuration.DefaultRedirectionURL != "" {
			err := ctx.SetJSONBody(redirectResponse{Redirect: ctx.Configuration.DefaultRedirectionURL})
			if err != nil {
				ctx.Logger.Errorf("Unable to set default redirection URL in body: %s", err)
			}
		} else {
			ctx.ReplyOK()
		}

		return
	}

	valid, err := utils.IsRedirectionURISafe(targetURI, ctx.Configuration.Session.Domain)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to check target URL: %s", err), mfaValidationFailedMessage)
		return
	}

	if valid {
		ctx.Logger.Debugf("Redirection URL %s is safe", targetURI)
		err := ctx.SetJSONBody(redirectResponse{Redirect: targetURI})

		if err != nil {
			ctx.Logger.Errorf("Unable to set redirection URL in body: %s", err)
		}
	} else {
		ctx.ReplyOK()
	}
}

// handleAuthenticationUnauthorized provides harmonized response codes for 1FA.
func handleAuthenticationUnauthorized(ctx *middlewares.AutheliaCtx, err error, message string) {
	ctx.SetStatusCode(fasthttp.StatusUnauthorized)
	ctx.Error(err, message)
}
