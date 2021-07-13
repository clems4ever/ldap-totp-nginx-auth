package handlers

import (
	"net/url"

	"github.com/authelia/authelia/internal/duo"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/utils"
)

// DuoPreAuth helper function for retrieving supported devices and capabilities from duo api.
func DuoPreAuth(duoAPI duo.API, ctx *middlewares.AutheliaCtx) (string, string, []DuoDevice, string, error) {
	userSession := ctx.GetSession()
	values := url.Values{}
	values.Set("username", userSession.Username)

	preauthResponse, err := duoAPI.PreauthCall(values, ctx)
	if err != nil {
		return "", "", nil, "", err
	}

	if preauthResponse.Result == auth {
		var supportedDevices []DuoDevice

		for _, device := range preauthResponse.Devices {
			var supportedMethods []string

			for _, method := range duo.PossibleMethods {
				if utils.IsStringInSlice(method, device.Capabilities) {
					supportedMethods = append(supportedMethods, method)
				}
			}

			if len(supportedMethods) > 0 {
				supportedDevices = append(supportedDevices, DuoDevice{
					Device:       device.Device,
					DisplayName:  device.DisplayName,
					Capabilities: supportedMethods,
				})
			}
		}

		if len(supportedDevices) > 0 {
			return preauthResponse.Result, preauthResponse.StatusMessage, supportedDevices, preauthResponse.EnrollPortalURL, nil
		}
	}

	return preauthResponse.Result, preauthResponse.StatusMessage, nil, preauthResponse.EnrollPortalURL, nil
}
