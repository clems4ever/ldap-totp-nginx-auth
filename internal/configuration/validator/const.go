package validator

import "regexp"

// Policy constants.
const (
	policyBypass    = "bypass"
	policyOneFactor = "one_factor"
	policyTwoFactor = "two_factor"
	policyDeny      = "deny"
)

// Hashing constants.
const (
	hashArgon2id = "argon2id"
	hashSHA512   = "sha512"
)

// Scheme constants.
const (
	schemeLDAP  = "ldap"
	schemeLDAPS = "ldaps"
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

// Test constants.
const (
	testBadTimer      = "-1"
	testInvalidPolicy = "invalid"
	testJWTSecret     = "a_secret"
	testLDAPBaseDN    = "base_dn"
	testLDAPPassword  = "password"
	testLDAPURL       = "ldap://ldap"
	testLDAPUser      = "user"
	testModeDisabled  = "disable"
	testTLSCert       = "/tmp/cert.pem"
	testTLSKey        = "/tmp/key.pem"
)

// Error constants.
const (
	errFmtDeprecatedConfigurationKey = "the %s configuration option is deprecated and will be " +
		"removed in %s, please use %s instead"
	errFmtReplacedConfigurationKey = "invalid configuration key '%s' was replaced by '%s'"

	errFmtLoggingLevelInvalid = "the log level '%s' is invalid, must be one of: %s"

	errFmtSessionSecretRedisProvider      = "the session secret must be set when using the %s session provider"
	errFmtSessionRedisPortRange           = "the port must be between 1 and 65535 for the %s session provider"
	errFmtSessionRedisHostRequired        = "the host must be provided when using the %s session provider"
	errFmtSessionRedisHostOrNodesRequired = "either the host or a node must be provided when using the %s session provider"

	errFmtOIDCClientInvalidSecret = "openid connect provider: client with ID '%s' has an empty secret"

	errFmtOIDCClientRedirectURI = "openid connect provider: client with ID '%s' redirect URI %s has an " +
		"invalid scheme %s, should be http or https"
	errFmtOIDCClientRedirectURICantBeParsed = "openid connect provider: client with ID '%s' has an invalid redirect " +
		"URI '%s' could not be parsed: %v"
	errFmtOIDCClientInvalidPolicy = "openid connect provider: client with ID '%s' has an invalid policy " +
		"'%s', should be either 'one_factor' or 'two_factor'"
	errFmtOIDCClientInvalidScope = "openid connect provider: client with ID '%s' has an invalid scope " +
		"'%s', must be one of: '%s'"
	errFmtOIDCClientInvalidGrantType = "openid connect provider: client with ID '%s' has an invalid grant type " +
		"'%s', must be one of: '%s'"
	errFmtOIDCClientInvalidResponseMode = "openid connect provider: client with ID '%s' has an invalid response mode " +
		"'%s', must be one of: '%s'"
	errFmtOIDCClientInvalidUserinfoAlgorithm = "openid connect provider: client with ID '%s' has an invalid userinfo signing " +
		"algorithm '%s', must be one of: '%s'"
	errFmtOIDCServerInsecureParameterEntropy = "SECURITY ISSUE: openid connect provider: minimum parameter entropy is " +
		"configured to an unsafe value, it should be above 8 but it's configured to %d."

	errFileHashing  = "config key incorrect: authentication_backend.file.hashing should be authentication_backend.file.password"
	errFilePHashing = "config key incorrect: authentication_backend.file.password_hashing should be authentication_backend.file.password"
	errFilePOptions = "config key incorrect: authentication_backend.file.password_options should be authentication_backend.file.password"

	errAccessControlInvalidPolicyWithSubjects = "policy [bypass] for rule #%d domain %s with subjects %s is invalid. It is " +
		"not supported to configure both policy bypass and subjects. For more information see: " +
		"https://www.authelia.com/docs/configuration/access-control.html#combining-subjects-and-the-bypass-policy"
)

var validLoggingLevels = []string{"trace", "debug", "info", "warn", "error"}
var validHTTPRequestMethods = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "TRACE", "CONNECT", "OPTIONS"}

var validOIDCScopes = []string{"openid", "email", "profile", "groups", "offline_access"}
var validOIDCGrantTypes = []string{"implicit", "refresh_token", "authorization_code", "password", "client_credentials"}
var validOIDCResponseModes = []string{"form_post", "query", "fragment"}
var validOIDCUserinfoAlgorithms = []string{"none", "RS256"}

var reKeyReplacer = regexp.MustCompile(`\[\d+]`)

// ValidKeys is a list of valid keys that are not secret names. For the sake of consistency please place any secret in
// the secret names map and reuse it in relevant sections.
var ValidKeys = []string{
	// Root Keys.
	"default_redirection_url",
	"jwt_secret",
	"theme",
	"certificates_directory",

	// Log keys.
	"log.level",
	"log.format",
	"log.file_path",
	"log.keep_stdout",

	// TODO: DEPRECATED START. Remove in 4.33.0.
	"host",
	"port",
	"tls_key",
	"tls_cert",
	"log_level",
	"log_format",
	"log_file_path",
	// TODO: DEPRECATED END. Remove in 4.33.0.

	// Server Keys.
	"server.host",
	"server.port",
	"server.tls_cert",
	"server.tls_key",
	"server.read_buffer_size",
	"server.write_buffer_size",
	"server.path",
	"server.enable_pprof",
	"server.enable_expvars",

	// TOTP Keys.
	"totp.issuer",
	"totp.period",
	"totp.skew",

	// DUO API Keys.
	"duo_api.hostname",
	"duo_api.secret_key",
	"duo_api.integration_key",

	// Access Control Keys.
	"access_control.default_policy",
	"access_control.networks",
	"access_control.rules",
	"access_control.rules[].domain",
	"access_control.rules[].methods",
	"access_control.rules[].networks",
	"access_control.rules[].subject",
	"access_control.rules[].policy",
	"access_control.rules[].resources",

	// Session Keys.
	"session.name",
	"session.domain",
	"session.secret",
	"session.same_site",
	"session.expiration",
	"session.inactivity",
	"session.remember_me_duration",

	// Redis Session Keys.
	"session.redis.host",
	"session.redis.port",
	"session.redis.username",
	"session.redis.password",
	"session.redis.database_index",
	"session.redis.maximum_active_connections",
	"session.redis.minimum_idle_connections",
	"session.redis.tls.minimum_version",
	"session.redis.tls.skip_verify",
	"session.redis.tls.server_name",
	"session.redis.high_availability.sentinel_name",
	"session.redis.high_availability.sentinel_password",
	"session.redis.high_availability.nodes",
	"session.redis.high_availability.route_by_latency",
	"session.redis.high_availability.route_randomly",
	"session.redis.timeouts.dial",
	"session.redis.timeouts.idle",
	"session.redis.timeouts.pool",
	"session.redis.timeouts.read",
	"session.redis.timeouts.write",

	// Local Storage Keys.
	"storage.local.path",

	// MySQL Storage Keys.
	"storage.mysql.host",
	"storage.mysql.port",
	"storage.mysql.database",
	"storage.mysql.username",
	"storage.mysql.password",

	// PostgreSQL Storage Keys.
	"storage.postgres.host",
	"storage.postgres.port",
	"storage.postgres.database",
	"storage.postgres.username",
	"storage.postgres.password",
	"storage.postgres.sslmode",

	// FileSystem Notifier Keys.
	"notifier.filesystem.filename",
	"notifier.disable_startup_check",

	// SMTP Notifier Keys.
	"notifier.smtp.host",
	"notifier.smtp.port",
	"notifier.smtp.username",
	"notifier.smtp.password",
	"notifier.smtp.identifier",
	"notifier.smtp.sender",
	"notifier.smtp.subject",
	"notifier.smtp.startup_check_address",
	"notifier.smtp.disable_require_tls",
	"notifier.smtp.disable_html_emails",
	"notifier.smtp.tls.minimum_version",
	"notifier.smtp.tls.skip_verify",
	"notifier.smtp.tls.server_name",

	// Regulation Keys.
	"regulation.max_retries",
	"regulation.find_time",
	"regulation.ban_time",

	// Authentication Backend Keys.
	"authentication_backend.disable_reset_password",
	"authentication_backend.refresh_interval",

	// LDAP Authentication Backend Keys.
	"authentication_backend.ldap.implementation",
	"authentication_backend.ldap.url",
	"authentication_backend.ldap.base_dn",
	"authentication_backend.ldap.username_attribute",
	"authentication_backend.ldap.additional_users_dn",
	"authentication_backend.ldap.users_filter",
	"authentication_backend.ldap.additional_groups_dn",
	"authentication_backend.ldap.groups_filter",
	"authentication_backend.ldap.group_name_attribute",
	"authentication_backend.ldap.mail_attribute",
	"authentication_backend.ldap.display_name_attribute",
	"authentication_backend.ldap.user",
	"authentication_backend.ldap.password",
	"authentication_backend.ldap.start_tls",
	"authentication_backend.ldap.tls.minimum_version",
	"authentication_backend.ldap.tls.skip_verify",
	"authentication_backend.ldap.tls.server_name",

	// File Authentication Backend Keys.
	"authentication_backend.file.path",
	"authentication_backend.file.password.algorithm",
	"authentication_backend.file.password.iterations",
	"authentication_backend.file.password.key_length",
	"authentication_backend.file.password.salt_length",
	"authentication_backend.file.password.memory",
	"authentication_backend.file.password.parallelism",

	// Identity Provider Keys.
	"identity_providers.oidc.hmac_secret",
	"identity_providers.oidc.issuer_private_key",
	"identity_providers.oidc.id_token_lifespan",
	"identity_providers.oidc.access_token_lifespan",
	"identity_providers.oidc.refresh_token_lifespan",
	"identity_providers.oidc.authorize_code_lifespan",
	"identity_providers.oidc.enable_client_debug_messages",
	"identity_providers.oidc.clients",
	"identity_providers.oidc.clients[].id",
	"identity_providers.oidc.clients[].description",
	"identity_providers.oidc.clients[].secret",
	"identity_providers.oidc.clients[].redirect_uris",
	"identity_providers.oidc.clients[].authorization_policy",
	"identity_providers.oidc.clients[].scopes",
	"identity_providers.oidc.clients[].grant_types",
	"identity_providers.oidc.clients[].response_types",
}

var replacedKeys = map[string]string{
	"authentication_backend.ldap.skip_verify":         "authentication_backend.ldap.tls.skip_verify",
	"authentication_backend.ldap.minimum_tls_version": "authentication_backend.ldap.tls.minimum_version",
	"notifier.smtp.disable_verify_cert":               "notifier.smtp.tls.skip_verify",
	"logs_file_path":                                  "log.file_path",
	"logs_level":                                      "log.level",
}

var specificErrorKeys = map[string]string{
	"google_analytics": "config key removed: google_analytics - this functionality has been deprecated",
	"notifier.smtp.trusted_cert": "invalid configuration key `notifier.smtp.trusted_cert` it has been removed, " +
		"option has been replaced by the global option `certificates_directory`",

	"authentication_backend.file.password_options.algorithm":   errFilePOptions,
	"authentication_backend.file.password_options.iterations":  errFilePOptions,
	"authentication_backend.file.password_options.key_length":  errFilePOptions,
	"authentication_backend.file.password_options.salt_length": errFilePOptions,
	"authentication_backend.file.password_options.memory":      errFilePOptions,
	"authentication_backend.file.password_options.parallelism": errFilePOptions,
	"authentication_backend.file.password_hashing.algorithm":   errFilePHashing,
	"authentication_backend.file.password_hashing.iterations":  errFilePHashing,
	"authentication_backend.file.password_hashing.key_length":  errFilePHashing,
	"authentication_backend.file.password_hashing.salt_length": errFilePHashing,
	"authentication_backend.file.password_hashing.memory":      errFilePHashing,
	"authentication_backend.file.password_hashing.parallelism": errFilePHashing,
	"authentication_backend.file.hashing.algorithm":            errFileHashing,
	"authentication_backend.file.hashing.iterations":           errFileHashing,
	"authentication_backend.file.hashing.key_length":           errFileHashing,
	"authentication_backend.file.hashing.salt_length":          errFileHashing,
	"authentication_backend.file.hashing.memory":               errFileHashing,
	"authentication_backend.file.hashing.parallelism":          errFileHashing,
}
