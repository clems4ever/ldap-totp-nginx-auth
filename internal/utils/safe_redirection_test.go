package utils

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func isURLSafe(requestURI string, domain string) bool { //nolint:unparam
	url, _ := url.ParseRequestURI(requestURI)
	return IsRedirectionSafe(*url, domain)
}

func TestIsRedirectionSafe_ShouldReturnFalseOnBadScheme(t *testing.T) {
	assert.False(t, isURLSafe("http://secure.example.com", "example.com"))
	assert.False(t, isURLSafe("ftp://secure.example.com", "example.com"))
	assert.True(t, isURLSafe("https://secure.example.com", "example.com"))
}

func TestIsRedirectionSafe_ShouldReturnFalseOnBadDomain(t *testing.T) {
	assert.False(t, isURLSafe("https://secure.example.com.c", "example.com"))
	assert.False(t, isURLSafe("https://secure.example.comc", "example.com"))
	assert.False(t, isURLSafe("https://secure.example.co", "example.com"))
}

func TestIsRedirectionURISafe_CannotParseURI(t *testing.T) {
	_, err := IsRedirectionURISafe("http//invalid", "example.com")
	assert.EqualError(t, err, "Unable to parse redirection URI http//invalid: parse \"http//invalid\": invalid URI for request")
}

func TestIsRedirectionURISafe_InvalidRedirectionURI(t *testing.T) {
	valid, err := IsRedirectionURISafe("http://myurl.com/myresource", "example.com")
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestIsRedirectionURISafe_ValidRedirectionURI(t *testing.T) {
	valid, err := IsRedirectionURISafe("http://myurl.example.com/myresource", "example.com")
	assert.NoError(t, err)
	assert.False(t, valid)
}
