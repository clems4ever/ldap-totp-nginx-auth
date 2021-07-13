package schema

import (
	"time"
)

const argon2id = "argon2id"

const (
	// ProfileRefreshDisabled represents a value for refresh_interval that disables the check entirely.
	ProfileRefreshDisabled = "disable"

	// ProfileRefreshAlways represents a value for refresh_interval that's the same as 0ms.
	ProfileRefreshAlways = "always"
)

const (
	// RefreshIntervalDefault represents the default value of refresh_interval.
	RefreshIntervalDefault = "5m"

	// RefreshIntervalAlways represents the duration value refresh interval should have if set to always.
	RefreshIntervalAlways = 0 * time.Millisecond
)

const (
	// LDAPImplementationCustom is the string for the custom LDAP implementation.
	LDAPImplementationCustom = "custom"

	// LDAPImplementationActiveDirectory is the string for the Active Directory LDAP implementation.
	LDAPImplementationActiveDirectory = "activedirectory"

	// LDAPImplementationFreeIPA is the string for the FreeIPA LDAP implementation.
	LDAPImplementationFreeIPA = "freeipa"
)
