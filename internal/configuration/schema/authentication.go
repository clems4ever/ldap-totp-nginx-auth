package schema

// LDAPAuthenticationBackendConfiguration represents the configuration related to LDAP server.
type LDAPAuthenticationBackendConfiguration struct {
	Implementation             string     `mapstructure:"implementation"`
	URL                        string     `mapstructure:"url"`
	BaseDN                     string     `mapstructure:"base_dn"`
	AdditionalUsersDN          string     `mapstructure:"additional_users_dn"`
	UsersFilter                string     `mapstructure:"users_filter"`
	AdditionalGroupsDN         string     `mapstructure:"additional_groups_dn"`
	GroupsFilter               string     `mapstructure:"groups_filter"`
	GroupNameAttribute         string     `mapstructure:"group_name_attribute"`
	UsernameAttribute          string     `mapstructure:"username_attribute"`
	MailAttribute              string     `mapstructure:"mail_attribute"`
	DisplayNameAttribute       string     `mapstructure:"display_name_attribute"`
	DistinguishedNameAttribute string     `mapstructure:"distinguished_name_attribute"`
	GroupsAttribute            string     `mapstructure:"groups_attribute"`
	User                       string     `mapstructure:"user"`
	Password                   string     `mapstructure:"password"`
	StartTLS                   bool       `mapstructure:"start_tls"`
	TLS                        *TLSConfig `mapstructure:"tls"`
}

// FileAuthenticationBackendConfiguration represents the configuration related to file-based backend.
type FileAuthenticationBackendConfiguration struct {
	Path     string                 `mapstructure:"path"`
	Password *PasswordConfiguration `mapstructure:"password"`
}

// PasswordConfiguration represents the configuration related to password hashing.
type PasswordConfiguration struct {
	Iterations  int    `mapstructure:"iterations"`
	KeyLength   int    `mapstructure:"key_length"`
	SaltLength  int    `mapstructure:"salt_length"`
	Algorithm   string `mapstrucutre:"algorithm"`
	Memory      int    `mapstructure:"memory"`
	Parallelism int    `mapstructure:"parallelism"`
}

// AuthenticationBackendConfiguration represents the configuration related to the authentication backend.
type AuthenticationBackendConfiguration struct {
	DisableResetPassword bool                                    `mapstructure:"disable_reset_password"`
	RefreshInterval      string                                  `mapstructure:"refresh_interval"`
	LDAP                 *LDAPAuthenticationBackendConfiguration `mapstructure:"ldap"`
	File                 *FileAuthenticationBackendConfiguration `mapstructure:"file"`
}

// DefaultPasswordConfiguration represents the default configuration related to Argon2id hashing.
var DefaultPasswordConfiguration = PasswordConfiguration{
	Iterations:  1,
	KeyLength:   32,
	SaltLength:  16,
	Algorithm:   argon2id,
	Memory:      64,
	Parallelism: 8,
}

// DefaultCIPasswordConfiguration represents the default configuration related to Argon2id hashing for CI.
var DefaultCIPasswordConfiguration = PasswordConfiguration{
	Iterations:  1,
	KeyLength:   32,
	SaltLength:  16,
	Algorithm:   argon2id,
	Memory:      64,
	Parallelism: 8,
}

// DefaultPasswordSHA512Configuration represents the default configuration related to SHA512 hashing.
var DefaultPasswordSHA512Configuration = PasswordConfiguration{
	Iterations: 50000,
	SaltLength: 16,
	Algorithm:  "sha512",
}

// DefaultLDAPAuthenticationBackendConfiguration represents the default LDAP config.
var DefaultLDAPAuthenticationBackendConfiguration = LDAPAuthenticationBackendConfiguration{
	Implementation:             LDAPImplementationCustom,
	UsernameAttribute:          "uid",
	MailAttribute:              "mail",
	DisplayNameAttribute:       "displayname",
	DistinguishedNameAttribute: "dn",
	GroupNameAttribute:         "cn",
	TLS: &TLSConfig{
		MinimumVersion: "TLS1.2",
	},
}

// DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration represents the default LDAP config for the MSAD Implementation.
var DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration = LDAPAuthenticationBackendConfiguration{
	UsersFilter:                "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2)(!pwdLastSet=0))",
	UsernameAttribute:          "sAMAccountName",
	MailAttribute:              "mail",
	DisplayNameAttribute:       "displayName",
	DistinguishedNameAttribute: "distinguishedName",
	GroupsFilter:               "(&(member={dn})(objectClass=group))",
	GroupNameAttribute:         "cn",
}

// DefaultLDAPAuthenticationBackendImplementationFreeIPAConfiguration represents the default LDAP config for the MSAD Implementation.
var DefaultLDAPAuthenticationBackendImplementationFreeIPAConfiguration = LDAPAuthenticationBackendConfiguration{
	UsersFilter:                "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=inetorgperson))",
	UsernameAttribute:          "uid",
	MailAttribute:              "mail",
	DisplayNameAttribute:       "displayName",
	DistinguishedNameAttribute: "dn",
	GroupsAttribute:            "memberOf",
	GroupNameAttribute:         "cn",
}
