package duo

import (
	"encoding/json"
	"net/url"

	duoapi "github.com/duosecurity/duo_api_golang"

	"github.com/authelia/authelia/internal/middlewares"
)

// API interface wrapping duo api library for testing purpose.
type API interface {
	Call(values url.Values, ctx *middlewares.AutheliaCtx, method string, path string) (*Response, error)
	PreauthCall(values url.Values, ctx *middlewares.AutheliaCtx) (*PreauthResponse, error)
	AuthCall(values url.Values, ctx *middlewares.AutheliaCtx) (*AuthResponse, error)
}

// APIImpl implementation of DuoAPI interface.
type APIImpl struct {
	*duoapi.DuoApi
}

// Device holds all necessary info for frontend.
type Device struct {
	Capabilities []string `json:"capabilities"`
	Device       string   `json:"device"`
	DisplayName  string   `json:"display_name"`
	Name         string   `json:"name"`
	SmsNextcode  string   `json:"sms_nextcode"`
	Number       string   `json:"number"`
	Type         string   `json:"type"`
}

// Response coming from Duo API.
type Response struct {
	Response      json.RawMessage `json:"response"`
	Code          int             `json:"code"`
	Message       string          `json:"message"`
	MessageDetail string          `json:"message_detail"`
	Stat          string          `json:"stat"`
}

// AuthResponse is a response for a authorization request.
type AuthResponse struct {
	Result             string `json:"result"`
	Status             string `json:"status"`
	StatusMessage      string `json:"status_msg"`
	TrustedDeviceToken string `json:"trusted_device_token"`
}

// PreauthResponse is a response for a preauthorization request.
type PreauthResponse struct {
	Result          string   `json:"result"`
	StatusMessage   string   `json:"status_msg"`
	Devices         []Device `json:"devices"`
	EnrollPortalURL string   `json:"enroll_portal_url"`
}
