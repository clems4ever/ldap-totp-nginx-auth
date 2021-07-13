package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldRaiseErrorWhenBothBackendsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	backendConfig := schema.AuthenticationBackendConfiguration{}

	backendConfig.LDAP = &schema.LDAPAuthenticationBackendConfiguration{}
	backendConfig.File = &schema.FileAuthenticationBackendConfiguration{
		Path: "/tmp",
	}

	ValidateAuthenticationBackend(&backendConfig, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "You cannot provide both `ldap` and `file` objects in `authentication_backend`")
}

func TestShouldRaiseErrorWhenNoBackendProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	backendConfig := schema.AuthenticationBackendConfiguration{}

	ValidateAuthenticationBackend(&backendConfig, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Please provide `ldap` or `file` object in `authentication_backend`")
}

type FileBasedAuthenticationBackend struct {
	suite.Suite
	configuration schema.AuthenticationBackendConfiguration
	validator     *schema.StructValidator
}

func (suite *FileBasedAuthenticationBackend) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration = schema.AuthenticationBackendConfiguration{}
	suite.configuration.File = &schema.FileAuthenticationBackendConfiguration{Path: "/a/path", Password: &schema.PasswordConfiguration{
		Algorithm:   schema.DefaultPasswordConfiguration.Algorithm,
		Iterations:  schema.DefaultPasswordConfiguration.Iterations,
		Parallelism: schema.DefaultPasswordConfiguration.Parallelism,
		Memory:      schema.DefaultPasswordConfiguration.Memory,
		KeyLength:   schema.DefaultPasswordConfiguration.KeyLength,
		SaltLength:  schema.DefaultPasswordConfiguration.SaltLength,
	}}
	suite.configuration.File.Password.Algorithm = schema.DefaultPasswordConfiguration.Algorithm
}
func (suite *FileBasedAuthenticationBackend) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenNoPathProvided() {
	suite.configuration.File.Path = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a `path` for the users database in `authentication_backend`")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenMemoryNotMoreThanEightTimesParallelism() {
	suite.configuration.File.Password.Memory = 8
	suite.configuration.File.Password.Parallelism = 2

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Memory for argon2id must be 16 or more (parallelism * 8), you configured memory as 8 and parallelism as 2")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultConfigurationWhenBlank() {
	suite.configuration.File.Password = &schema.PasswordConfiguration{}

	suite.Assert().Equal(0, suite.configuration.File.Password.KeyLength)
	suite.Assert().Equal(0, suite.configuration.File.Password.Iterations)
	suite.Assert().Equal(0, suite.configuration.File.Password.SaltLength)
	suite.Assert().Equal("", suite.configuration.File.Password.Algorithm)
	suite.Assert().Equal(0, suite.configuration.File.Password.Memory)
	suite.Assert().Equal(0, suite.configuration.File.Password.Parallelism)

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.DefaultPasswordConfiguration.KeyLength, suite.configuration.File.Password.KeyLength)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Iterations, suite.configuration.File.Password.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.SaltLength, suite.configuration.File.Password.SaltLength)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Algorithm, suite.configuration.File.Password.Algorithm)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Memory, suite.configuration.File.Password.Memory)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Parallelism, suite.configuration.File.Password.Parallelism)
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultConfigurationWhenOnlySHA512Set() {
	suite.configuration.File.Password = &schema.PasswordConfiguration{}
	suite.Assert().Equal("", suite.configuration.File.Password.Algorithm)
	suite.configuration.File.Password.Algorithm = "sha512"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.KeyLength, suite.configuration.File.Password.KeyLength)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Iterations, suite.configuration.File.Password.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.SaltLength, suite.configuration.File.Password.SaltLength)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Algorithm, suite.configuration.File.Password.Algorithm)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Memory, suite.configuration.File.Password.Memory)
	suite.Assert().Equal(schema.DefaultPasswordSHA512Configuration.Parallelism, suite.configuration.File.Password.Parallelism)
}
func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenKeyLengthTooLow() {
	suite.configuration.File.Password.KeyLength = 1

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Key length for argon2id must be 16, you configured 1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenSaltLengthTooLow() {
	suite.configuration.File.Password.SaltLength = -1

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "The salt length must be 2 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenBadAlgorithmDefined() {
	suite.configuration.File.Password.Algorithm = "bogus"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Unknown hashing algorithm supplied, valid values are argon2id and sha512, you configured 'bogus'")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenIterationsTooLow() {
	suite.configuration.File.Password.Iterations = -1

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "The number of iterations specified is invalid, must be 1 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldRaiseErrorWhenParallelismTooLow() {
	suite.configuration.File.Password.Parallelism = -1

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Parallelism for argon2id must be 1 or more, you configured -1")
}

func (suite *FileBasedAuthenticationBackend) TestShouldSetDefaultValues() {
	suite.configuration.File.Password.Algorithm = ""
	suite.configuration.File.Password.Iterations = 0
	suite.configuration.File.Password.SaltLength = 0
	suite.configuration.File.Password.Memory = 0
	suite.configuration.File.Password.Parallelism = 0

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Algorithm, suite.configuration.File.Password.Algorithm)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Iterations, suite.configuration.File.Password.Iterations)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.SaltLength, suite.configuration.File.Password.SaltLength)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Memory, suite.configuration.File.Password.Memory)
	suite.Assert().Equal(schema.DefaultPasswordConfiguration.Parallelism, suite.configuration.File.Password.Parallelism)
}

func TestFileBasedAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(FileBasedAuthenticationBackend))
}

type LDAPAuthenticationBackendSuite struct {
	suite.Suite
	configuration schema.AuthenticationBackendConfiguration
	validator     *schema.StructValidator
}

func (suite *LDAPAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration = schema.AuthenticationBackendConfiguration{}
	suite.configuration.LDAP = &schema.LDAPAuthenticationBackendConfiguration{}
	suite.configuration.LDAP.Implementation = schema.LDAPImplementationCustom
	suite.configuration.LDAP.URL = testLDAPURL
	suite.configuration.LDAP.User = testLDAPUser
	suite.configuration.LDAP.Password = testLDAPPassword
	suite.configuration.LDAP.BaseDN = testLDAPBaseDN
	suite.configuration.LDAP.UsernameAttribute = "uid"
	suite.configuration.LDAP.UsersFilter = "({username_attribute}={input})"
	suite.configuration.LDAP.GroupsFilter = "(cn={input})"
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateCompleteConfiguration() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldValidateDefaultImplementationAndUsernameAttribute() {
	suite.configuration.LDAP.Implementation = ""
	suite.configuration.LDAP.UsernameAttribute = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().Equal(schema.LDAPImplementationCustom, suite.configuration.LDAP.Implementation)

	suite.Assert().Equal(suite.configuration.LDAP.UsernameAttribute, schema.DefaultLDAPAuthenticationBackendConfiguration.UsernameAttribute)
	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenImplementationIsInvalidMSAD() {
	suite.configuration.LDAP.Implementation = "masd"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication backend ldap implementation must be blank or one of the following values `custom`, `activedirectory`")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenURLNotProvided() {
	suite.configuration.LDAP.URL = ""
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a URL to the LDAP server")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenUserNotProvided() {
	suite.configuration.LDAP.User = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a user name to connect to the LDAP server")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenPasswordNotProvided() {
	suite.configuration.LDAP.Password = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a password to connect to the LDAP server")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorWhenBaseDNNotProvided() {
	suite.configuration.LDAP.BaseDN = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a base DN to connect to the LDAP server")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnEmptyGroupsFilter() {
	suite.configuration.LDAP.GroupsFilter = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a groups filter with `groups_filter` attribute")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnEmptyUsersFilter() {
	suite.configuration.LDAP.UsersFilter = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Please provide a users filter with `users_filter` attribute")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotRaiseOnEmptyUsernameAttribute() {
	suite.configuration.LDAP.UsernameAttribute = ""

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseOnBadRefreshInterval() {
	suite.configuration.RefreshInterval = "blah"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Auth Backend `refresh_interval` is configured to 'blah' but it must be either a duration notation or one of 'disable', or 'always'. Error from parser: could not convert the input string of blah into a duration")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultImplementation() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.LDAPImplementationCustom, suite.configuration.LDAP.Implementation)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseErrorOnBadFilterPlaceholders() {
	suite.configuration.LDAP.UsersFilter = "(&({username_attribute}={0})(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
	suite.configuration.LDAP.GroupsFilter = "(&(member={0})(objectClass=group)(objectCategory=group))"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().True(suite.validator.HasErrors())

	suite.Require().Len(suite.validator.Errors(), 2)
	suite.Assert().EqualError(suite.validator.Errors()[0], "authentication backend ldap users filter must "+
		"not contain removed placeholders, {0} has been replaced with {input}")
	suite.Assert().EqualError(suite.validator.Errors()[1], "authentication backend ldap groups filter must "+
		"not contain removed placeholders, "+
		"{0} has been replaced with {input} and {1} has been replaced with {username}")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultGroupNameAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("cn", suite.configuration.LDAP.GroupNameAttribute)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultMailAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("mail", suite.configuration.LDAP.MailAttribute)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultDisplayNameAttribute() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("displayname", suite.configuration.LDAP.DisplayNameAttribute)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultRefreshInterval() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal("5m", suite.configuration.RefreshInterval)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainEnclosingParenthesis() {
	suite.configuration.LDAP.UsersFilter = "{username_attribute}={input}"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "The users filter should contain enclosing parenthesis. For instance {username_attribute}={input} should be ({username_attribute}={input})")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenGroupsFilterDoesNotContainEnclosingParenthesis() {
	suite.configuration.LDAP.GroupsFilter = "cn={input}"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "The groups filter should contain enclosing parenthesis. For instance cn={input} should be (cn={input})")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldRaiseWhenUsersFilterDoesNotContainUsernameAttribute() {
	suite.configuration.LDAP.UsersFilter = "(&({mail_attribute}={input})(objectClass=person))"
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Unable to detect {username_attribute} placeholder in users_filter, your configuration is broken. Please review configuration options listed at https://www.authelia.com/docs/configuration/authentication/ldap.html")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldHelpDetectNoInputPlaceholder() {
	suite.configuration.LDAP.UsersFilter = "(&({username_attribute}={mail_attribute})(objectClass=person))"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Unable to detect {input} placeholder in users_filter, your configuration might be broken. Please review configuration options listed at https://www.authelia.com/docs/configuration/authentication/ldap.html")
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldAdaptLDAPURL() {
	suite.Assert().Equal("", validateLDAPURLSimple(loopback, suite.validator))

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Unknown scheme for ldap url, should be ldap:// or ldaps://")

	suite.Assert().Equal("", validateLDAPURLSimple("127.0.0.1:636", suite.validator))

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 2)
	suite.Assert().EqualError(suite.validator.Errors()[1], "Unable to parse URL to ldap server. The scheme is probably missing: ldap:// or ldaps://")

	suite.Assert().Equal("ldap://127.0.0.1", validateLDAPURLSimple("ldap://127.0.0.1", suite.validator))
	suite.Assert().Equal("ldap://127.0.0.1:390", validateLDAPURLSimple("ldap://127.0.0.1:390", suite.validator))
	suite.Assert().Equal("ldap://127.0.0.1/abc", validateLDAPURLSimple("ldap://127.0.0.1/abc", suite.validator))
	suite.Assert().Equal("ldap://127.0.0.1/abc?test=abc&x=y", validateLDAPURLSimple("ldap://127.0.0.1/abc?test=abc&x=y", suite.validator))

	suite.Assert().Equal("ldaps://127.0.0.1:390", validateLDAPURLSimple("ldaps://127.0.0.1:390", suite.validator))
	suite.Assert().Equal("ldaps://127.0.0.1", validateLDAPURLSimple("ldaps://127.0.0.1", suite.validator))
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldSetDefaultTLSMinimumVersion() {
	suite.configuration.LDAP.TLS = &schema.TLSConfig{MinimumVersion: ""}

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(schema.DefaultLDAPAuthenticationBackendConfiguration.TLS.MinimumVersion, suite.configuration.LDAP.TLS.MinimumVersion)
}

func (suite *LDAPAuthenticationBackendSuite) TestShouldNotAllowInvalidTLSValue() {
	suite.configuration.LDAP.TLS = &schema.TLSConfig{
		MinimumVersion: "SSL2.0",
	}

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "error occurred validating the LDAP minimum_tls_version key with value SSL2.0: supplied TLS version isn't supported")
}

func TestLdapAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(LDAPAuthenticationBackendSuite))
}

type ActiveDirectoryAuthenticationBackendSuite struct {
	suite.Suite
	configuration schema.AuthenticationBackendConfiguration
	validator     *schema.StructValidator
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration = schema.AuthenticationBackendConfiguration{}
	suite.configuration.LDAP = &schema.LDAPAuthenticationBackendConfiguration{}
	suite.configuration.LDAP.Implementation = schema.LDAPImplementationActiveDirectory
	suite.configuration.LDAP.URL = testLDAPURL
	suite.configuration.LDAP.User = testLDAPUser
	suite.configuration.LDAP.Password = testLDAPPassword
	suite.configuration.LDAP.BaseDN = testLDAPBaseDN
	suite.configuration.LDAP.TLS = schema.DefaultLDAPAuthenticationBackendConfiguration.TLS
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldSetActiveDirectoryDefaults() {
	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())

	suite.Assert().Equal(
		suite.configuration.LDAP.UsersFilter,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsersFilter)
	suite.Assert().Equal(
		suite.configuration.LDAP.UsernameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsernameAttribute)
	suite.Assert().Equal(
		suite.configuration.LDAP.DisplayNameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.DisplayNameAttribute)
	suite.Assert().Equal(
		suite.configuration.LDAP.MailAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.MailAttribute)
	suite.Assert().Equal(
		suite.configuration.LDAP.GroupsFilter,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupsFilter)
	suite.Assert().Equal(
		suite.configuration.LDAP.GroupNameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupNameAttribute)
}

func (suite *ActiveDirectoryAuthenticationBackendSuite) TestShouldOnlySetDefaultsIfNotManuallyConfigured() {
	suite.configuration.LDAP.UsersFilter = "(&({username_attribute}={input})(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2))"
	suite.configuration.LDAP.UsernameAttribute = "cn"
	suite.configuration.LDAP.MailAttribute = "userPrincipalName"
	suite.configuration.LDAP.DisplayNameAttribute = "name"
	suite.configuration.LDAP.GroupsFilter = "(&(member={dn})(objectClass=group)(objectCategory=group))"
	suite.configuration.LDAP.GroupNameAttribute = "distinguishedName"

	ValidateAuthenticationBackend(&suite.configuration, suite.validator)

	suite.Assert().NotEqual(
		suite.configuration.LDAP.UsersFilter,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsersFilter)
	suite.Assert().NotEqual(
		suite.configuration.LDAP.UsernameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.UsernameAttribute)
	suite.Assert().NotEqual(
		suite.configuration.LDAP.DisplayNameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.DisplayNameAttribute)
	suite.Assert().NotEqual(
		suite.configuration.LDAP.MailAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.MailAttribute)
	suite.Assert().NotEqual(
		suite.configuration.LDAP.GroupsFilter,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupsFilter)
	suite.Assert().NotEqual(
		suite.configuration.LDAP.GroupNameAttribute,
		schema.DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration.GroupNameAttribute)
}

func TestActiveDirectoryAuthenticationBackend(t *testing.T) {
	suite.Run(t, new(ActiveDirectoryAuthenticationBackendSuite))
}
