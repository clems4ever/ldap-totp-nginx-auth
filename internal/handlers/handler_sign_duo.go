package handlers

import (
	"fmt"
	"net/url"

	"github.com/authelia/authelia/internal/duo"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/utils"
)

// SecondFactorDuoPost handler for sending a push notification via duo api.
func SecondFactorDuoPost(duoAPI duo.API) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var requestBody signDuoRequestBody

		if err := ctx.ParseBody(&requestBody); err != nil {
			handleAuthenticationUnauthorized(ctx, err, mfaValidationFailedMessage)
			return
		}

		userSession := ctx.GetSession()
		remoteIP := ctx.RemoteIP().String()

		device, method, err := ctx.Providers.StorageProvider.LoadPreferredDuoDevice(userSession.Username)
		if err != nil {
			ctx.Logger.Debugf("%s - Starting Duo PreAuth for %s", err, userSession.Username)

			device, method, err := HandlePreAuth(duoAPI, ctx, requestBody.TargetURL)
			if err != nil {
				ctx.Error(err, operationFailedMessage)
				return
			}

			if device == "" || method == "" {
				return
			}
		}

		ctx.Logger.Debugf("Starting Duo Auth Attempt for %s from IP %s", userSession.Username, remoteIP)

		if !utils.IsStringInSlice(method, duo.PossibleMethods) {
			ctx.Logger.Debugf("%s Preffered Duo method not supported: %s", method, userSession.Username)

			if err := ctx.SetJSONBody(DuoSignResponse{Result: auth}); err != nil {
				ctx.Error(fmt.Errorf("Unable to set JSON body in response"), operationFailedMessage)
			}

			return
		}

		values, err := SetValues(userSession, device, method, remoteIP, requestBody.TargetURL, requestBody.Passcode)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, err, mfaValidationFailedMessage)
			return
		}

		authResponse, err := duoAPI.AuthCall(values, ctx)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Duo API errored: %s", err), mfaValidationFailedMessage)
			return
		}

		if authResponse.Result != allow {
			ctx.ReplyUnauthorized()
			return
		}

		err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Unable to regenerate session for user %s: %s", userSession.Username, err), mfaValidationFailedMessage)
			return
		}

		userSession.SetTwoFactor(ctx.Clock.Now())

		err = ctx.SaveSession(userSession)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Unable to update authentication level with Duo: %s", err), mfaValidationFailedMessage)
			return
		}

		if userSession.OIDCWorkflowSession != nil {
			handleOIDCWorkflowResponse(ctx)
		} else {
			Handle2FAResponse(ctx, requestBody.TargetURL)
		}
	}
}

// HandlePreAuth handler for retrieving all available devices.
func HandlePreAuth(duoAPI duo.API, ctx *middlewares.AutheliaCtx, targetURL string) (string, string, error) {
	result, message, devices, enrollURL, err := DuoPreAuth(duoAPI, ctx)

	if err != nil {
		handleAuthenticationUnauthorized(ctx, fmt.Errorf("Duo PreAuth API errored: %s", err), mfaValidationFailedMessage)
		return "", "", nil
	}

	userSession := ctx.GetSession()

	if result == enroll {
		ctx.Logger.Debugf("Duo User not enrolled: %s", userSession.Username)

		if err := ctx.SetJSONBody(DuoSignResponse{Result: enroll, EnrollURL: enrollURL}); err != nil {
			return "", "", fmt.Errorf("Unable to set JSON body in response")
		}

		return "", "", nil
	}

	if result == deny {
		handleAuthenticationUnauthorized(ctx, fmt.Errorf("Duo User %s not allowed to authenticate: %s", userSession.Username, message), mfaValidationFailedMessage)
		return "", "", nil
	}

	if result == allow {
		ctx.Logger.Debugf("Duo authentication was bypassed for user %s", userSession.Username)
		Handle2FAResponse(ctx, targetURL)

		return "", "", nil
	}

	if result == auth {
		if devices == nil {
			ctx.Logger.Debugf("No applicable device/method available for Duo user %s", userSession.Username)

			if err := ctx.SetJSONBody(DuoSignResponse{Result: enroll}); err != nil {
				return "", "", fmt.Errorf("Unable to set JSON body in response")
			}

			return "", "", nil
		}

		if len(devices) > 1 {
			ctx.Logger.Debugf("Multiple devices available for Duo user %s - require selection", userSession.Username)

			if err := ctx.SetJSONBody(DuoSignResponse{Result: auth, Devices: devices}); err != nil {
				return "", "", fmt.Errorf("Unable to set JSON body in response")
			}

			return "", "", nil
		}

		if len(devices[0].Capabilities) > 1 {
			ctx.Logger.Debugf("Multiple methods available for Duo user %s - require selection", userSession.Username)

			if err := ctx.SetJSONBody(DuoSignResponse{Result: auth, Devices: devices}); err != nil {
				return "", "", fmt.Errorf("Unable to set JSON body in response")
			}

			return "", "", nil
		}

		device := devices[0].Device
		method := devices[0].Capabilities[0]
		ctx.Logger.Debugf("Exactly one device(%s) and method(%s) found - Saving as new preferred Duo device and method for user %s", device, method, userSession.Username)

		if err := ctx.Providers.StorageProvider.SavePreferredDuoDevice(userSession.Username, device, method); err != nil {
			return "", "", fmt.Errorf("Unable to save new preferred Duo device and method for user %s: %s", userSession.Username, err)
		}

		return device, method, nil
	}

	return "", "", fmt.Errorf("Unknown result code: %s", result)
}

// SetValues sets all appropriate Values for the Auth Request.
func SetValues(userSession session.UserSession, device string, method string, remoteIP string, targetURL string, passcode string) (url.Values, error) {
	values := url.Values{}
	values.Set("username", userSession.Username)
	values.Set("ipaddr", remoteIP)
	values.Set("factor", method)

	switch method {
	case duo.Push:
		values.Set("device", device)

		if userSession.DisplayName != "" {
			values.Set("display_username", userSession.DisplayName)
		}

		if targetURL != "" {
			values.Set("pushinfo", fmt.Sprintf("target%%20url=%s", targetURL))
		}
	case duo.Phone:
		values.Set("device", device)
	case duo.SMS:
		values.Set("device", device)
	case duo.OTP:
		if passcode != "" {
			values.Set("passcode", passcode)
		} else {
			return nil, fmt.Errorf("No Passcode received from user %s", userSession.Username)
		}
	}

	return values, nil
}
