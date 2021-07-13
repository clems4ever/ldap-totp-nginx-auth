package storage

import (
	"time"

	"github.com/authelia/authelia/internal/models"
)

// Provider is an interface providing storage capabilities for
// persisting any kind of data related to Authelia.
type Provider interface {
	LoadPreferred2FAMethod(username string) (string, error)
	SavePreferred2FAMethod(username string, method string) error

	FindIdentityVerificationToken(token string) (bool, error)
	SaveIdentityVerificationToken(token string) error
	RemoveIdentityVerificationToken(token string) error

	SaveTOTPSecret(username string, secret string) error
	LoadTOTPSecret(username string) (string, error)
	DeleteTOTPSecret(username string) error

	SaveU2FDeviceHandle(username string, keyHandle []byte, publicKey []byte) error
	LoadU2FDeviceHandle(username string) (keyHandle []byte, publicKey []byte, err error)

	SavePreferredDuoDevice(username string, device string, method string) error
	LoadPreferredDuoDevice(username string) (string, string, error)

	AppendAuthenticationLog(attempt models.AuthenticationAttempt) error
	LoadLatestAuthenticationLogs(username string, fromDate time.Time) ([]models.AuthenticationAttempt, error)
}
