package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/authelia/authelia/internal/duo"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/utils"
)

// SecondFactorDuoDevicesGet handler for retrieving available devices and capabilities from duo api.
func SecondFactorDuoDevicesGet(duoAPI duo.API) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		userSession := ctx.GetSession()
		values := url.Values{}
		values.Set("username", userSession.Username)

		ctx.Logger.Debugf("Starting Duo PreAuth for %s", userSession.Username)

		result, message, devices, enrollURL, err := DuoPreAuth(duoAPI, ctx)
		if err != nil {
			ctx.Error(fmt.Errorf("Duo PreAuth API errored: %s", err), operationFailedMessage)
			return
		}

		if result == auth {
			if devices == nil {
				ctx.Logger.Debugf("No applicable device/method available for Duo user %s", userSession.Username)

				if err := ctx.SetJSONBody(DuoDevicesResponse{Result: "enroll"}); err != nil {
					ctx.Error(fmt.Errorf("Unable to set JSON body in response"), operationFailedMessage)
				}

				return
			}

			if err := ctx.SetJSONBody(DuoDevicesResponse{Result: result, Devices: devices}); err != nil {
				ctx.Error(fmt.Errorf("Unable to set JSON body in response"), operationFailedMessage)
			}

			return
		}

		if result == allow {
			ctx.Logger.Debugf("Device selection not possible for user %s, because Duo authentication was bypassed - Defaults to Auto Push", userSession.Username)

			if err := ctx.SetJSONBody(DuoDevicesResponse{Result: result}); err != nil {
				ctx.Error(fmt.Errorf("Unable to set JSON body in response"), operationFailedMessage)
			}

			return
		}

		if result == enroll {
			ctx.Logger.Debugf("Duo User not enrolled: %s", userSession.Username)

			if err := ctx.SetJSONBody(DuoDevicesResponse{Result: result, EnrollURL: enrollURL}); err != nil {
				ctx.Error(fmt.Errorf("Unable to set JSON body in response"), operationFailedMessage)
			}

			return
		}

		if result == deny {
			ctx.Logger.Debugf("Duo User not allowed to authenticate: %s", userSession.Username)

			if err := ctx.SetJSONBody(DuoDevicesResponse{Result: result}); err != nil {
				ctx.Error(fmt.Errorf("Unable to set JSON body in response"), operationFailedMessage)
			}

			return
		}

		ctx.Error(fmt.Errorf("Duo PreAuth API errored for %s: %s - %s", userSession.Username, result, message), operationFailedMessage)
	}
}

// SecondFactorDuoDevicePost update the user preferences regarding Duo device and method.
func SecondFactorDuoDevicePost(ctx *middlewares.AutheliaCtx) {
	device := DuoDeviceBody{}

	err := ctx.ParseBody(&device)
	if err != nil {
		ctx.Error(err, operationFailedMessage)
		return
	}

	if !utils.IsStringInSlice(device.Method, duo.PossibleMethods) {
		ctx.Error(fmt.Errorf("Unknown method '%s', it should be one of %s", device.Method, strings.Join(duo.PossibleMethods, ", ")), operationFailedMessage)
		return
	}

	userSession := ctx.GetSession()
	ctx.Logger.Debugf("Save new preferred Duo device and method of user %s to %s using %s", userSession.Username, device.Device, device.Method)
	err = ctx.Providers.StorageProvider.SavePreferredDuoDevice(userSession.Username, device.Device, device.Method)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to save new preferred Duo device and method: %s", err), operationFailedMessage)
		return
	}

	ctx.ReplyOK()
}
