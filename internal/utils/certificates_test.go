package utils

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldReturnCorrectTLSVersions(t *testing.T) {
	tls13 := uint16(tls.VersionTLS13)
	tls12 := uint16(tls.VersionTLS12)
	tls11 := uint16(tls.VersionTLS11)
	tls10 := uint16(tls.VersionTLS10)

	version, err := TLSStringToTLSConfigVersion(TLS13)
	assert.Equal(t, tls13, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS13)
	assert.Equal(t, tls13, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(TLS12)
	assert.Equal(t, tls12, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS12)
	assert.Equal(t, tls12, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(TLS11)
	assert.Equal(t, tls11, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS11)
	assert.Equal(t, tls11, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion(TLS10)
	assert.Equal(t, tls10, version)
	assert.NoError(t, err)

	version, err = TLSStringToTLSConfigVersion("TLS" + TLS10)
	assert.Equal(t, tls10, version)
	assert.NoError(t, err)
}

func TestShouldReturnZeroAndErrorOnInvalidTLSVersions(t *testing.T) {
	version, err := TLSStringToTLSConfigVersion("TLS1.4")
	assert.Error(t, err)
	assert.Equal(t, uint16(0), version)
	assert.EqualError(t, err, "supplied TLS version isn't supported")

	version, err = TLSStringToTLSConfigVersion("SSL3.0")
	assert.Error(t, err)
	assert.Equal(t, uint16(0), version)
	assert.EqualError(t, err, "supplied TLS version isn't supported")
}

func TestShouldReturnErrWhenX509DirectoryNotExist(t *testing.T) {
	pool, errs, nonFatalErrs := NewX509CertPool("/tmp/asdfzyxabc123/not/a/real/dir", nil)
	assert.NotNil(t, pool)
	assert.Len(t, nonFatalErrs, 0)
	require.Len(t, errs, 1)
	assert.EqualError(t, errs[0], "could not read certificates from directory open /tmp/asdfzyxabc123/not/a/real/dir: no such file or directory")
}

func TestShouldNotReturnErrWhenX509DirectoryExist(t *testing.T) {
	pool, errs, nonFatalErrs := NewX509CertPool("/tmp", nil)
	assert.NotNil(t, pool)
	assert.Len(t, nonFatalErrs, 0)
	assert.Len(t, errs, 0)
}

func TestShouldRaiseNonFatalErrWhenNotifierTrustedCertConfigured(t *testing.T) {
	config := &schema.Configuration{
		Notifier: &schema.NotifierConfiguration{
			SMTP: &schema.SMTPNotifierConfiguration{
				TrustedCert: "../suites/common/ssl/cert.pem",
			},
		},
	}

	pool, errs, nonFatalErrs := NewX509CertPool("/tmp", config)
	assert.NotNil(t, pool)
	require.Len(t, nonFatalErrs, 1)
	assert.Len(t, errs, 0)

	assert.EqualError(t, nonFatalErrs[0], "defining the trusted cert in the SMTP notifier is deprecated and will be removed in 4.28.0, please use the global certificates_directory instead")
}

func TestShouldRaiseErrAndNonFatalErrWhenNotifierTrustedCertConfiguredAndNotExist(t *testing.T) {
	config := &schema.Configuration{
		Notifier: &schema.NotifierConfiguration{
			SMTP: &schema.SMTPNotifierConfiguration{
				TrustedCert: "/tmp/asdfzyxabc123/not/a/real/cert.pem",
			},
		},
	}

	pool, errs, nonFatalErrs := NewX509CertPool("/tmp", config)
	assert.NotNil(t, pool)
	require.Len(t, nonFatalErrs, 1)
	require.Len(t, errs, 1)

	assert.EqualError(t, errs[0], "could not import legacy SMTP trusted_cert (see the new certificates_directory option) certificate /tmp/asdfzyxabc123/not/a/real/cert.pem (file does not exist)")
	assert.EqualError(t, nonFatalErrs[0], "defining the trusted cert in the SMTP notifier is deprecated and will be removed in 4.28.0, please use the global certificates_directory instead")
}
