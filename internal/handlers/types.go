package handlers

import (
	"github.com/tstranex/u2f"

	"github.com/authelia/authelia/internal/authentication"
)

// MethodList is the list of available methods.
type MethodList = []string

type authorizationMatching int

// UserInfo is the model of user info and second factor preferences.
type UserInfo struct {
	// The users display name.
	DisplayName string `json:"display_name"`

	// The preferred 2FA method.
	Method string `json:"method" valid:"required"`

	// True if a security key has been registered.
	HasU2F bool `json:"has_u2f" valid:"required"`

	// True if a TOTP device has been registered.
	HasTOTP bool `json:"has_totp" valid:"required"`
}

// signTOTPRequestBody model of the request body received by TOTP authentication endpoint.
type signTOTPRequestBody struct {
	Token     string `json:"token" valid:"required"`
	TargetURL string `json:"targetURL"`
}

// signU2FRequestBody model of the request body of U2F authentication endpoint.
type signU2FRequestBody struct {
	SignResponse u2f.SignResponse `json:"signResponse"`
	TargetURL    string           `json:"targetURL"`
}

type signDuoRequestBody struct {
	TargetURL string `json:"targetURL"`
}

// firstFactorRequestBody represents the JSON body received by the endpoint.
type firstFactorRequestBody struct {
	Username       string `json:"username" valid:"required"`
	Password       string `json:"password" valid:"required"`
	TargetURL      string `json:"targetURL"`
	RequestMethod  string `json:"requestMethod"`
	KeepMeLoggedIn *bool  `json:"keepMeLoggedIn"`
	// KeepMeLoggedIn: Cannot require this field because of https://github.com/asaskevich/govalidator/pull/329
	// TODO(c.michaud): add required validation once the above PR is merged.
}

// checkURIWithinDomainRequestBody represents the JSON body received by the endpoint checking if an URI is within
// the configured domain.
type checkURIWithinDomainRequestBody struct {
	URI string `json:"uri"`
}

type checkURIWithinDomainResponseBody struct {
	OK bool `json:"ok"`
}

// redirectResponse represent the response sent by the first factor endpoint
// when a redirection URL has been provided.
type redirectResponse struct {
	Redirect string `json:"redirect"`
}

// TOTPKeyResponse is the model of response that is sent to the client up successful identity verification.
type TOTPKeyResponse struct {
	Base32Secret string `json:"base32_secret"`
	OTPAuthURL   string `json:"otpauth_url"`
}

// StateResponse represents the response sent by the state endpoint.
type StateResponse struct {
	Username              string               `json:"username"`
	AuthenticationLevel   authentication.Level `json:"authentication_level"`
	DefaultRedirectionURL string               `json:"default_redirection_url"`
}

// resetPasswordStep1RequestBody model of the reset password (step1) request body.
type resetPasswordStep1RequestBody struct {
	Username string `json:"username"`
}

// resetPasswordStep2RequestBody model of the reset password (step2) request body.
type resetPasswordStep2RequestBody struct {
	Password string `json:"password"`
}
