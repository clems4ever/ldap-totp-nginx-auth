package handlers

import (
	"fmt"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/utils"
)

// CheckSafeRedirection handler checking whether the redirection to a given URL provided in body is safe.
func CheckSafeRedirection(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()

	if userSession.AuthenticationLevel == authentication.NotAuthenticated {
		ctx.ReplyUnauthorized()
		return
	}

	var reqBody checkURIWithinDomainRequestBody

	err := ctx.ParseBody(&reqBody)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to parse request body: %w", err), operationFailedMessage)
		return
	}

	safe, err := utils.IsRedirectionURISafe(reqBody.URI, ctx.Configuration.Session.Domain)
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to determine if uri %s is safe to redirect to: %w", reqBody.URI, err), operationFailedMessage)
		return
	}

	err = ctx.SetJSONBody(checkURIWithinDomainResponseBody{
		OK: safe,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("Unable to create response body: %w", err), operationFailedMessage)
		return
	}
}
