package server

import (
	"net"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/logging"
)

// Replacement for the default error handler in fasthttp.
func autheliaErrorHandler(ctx *fasthttp.RequestCtx, err error) {
	logger := logging.Logger()

	if _, ok := err.(*fasthttp.ErrSmallBuffer); ok {
		// Note: Getting X-Forwarded-For or Request URI is impossible for ths error.
		logger.Tracef("request was too large to handle from client %s. response code %d.", ctx.RemoteIP().String(), fasthttp.StatusRequestHeaderFieldsTooLarge)
		ctx.Error("request header too large", fasthttp.StatusRequestHeaderFieldsTooLarge)
	} else if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
		// TODO: Add X-Forwarded-For Check here.
		logger.Tracef("request timeout occurred while handling from client %s: %s. response code %d.", ctx.RemoteIP().String(), ctx.RequestURI(), fasthttp.StatusRequestTimeout)
		ctx.Error("request timeout", fasthttp.StatusRequestTimeout)
	} else {
		// TODO: Add X-Forwarded-For Check here.
		logger.Tracef("an unknown error occurred while handling a request from client %s: %s. response code %d.", ctx.RemoteIP().String(), ctx.RequestURI(), fasthttp.StatusBadRequest)
		ctx.Error("error when parsing request", fasthttp.StatusBadRequest)
	}
}
