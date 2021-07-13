package authentication

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/unicode"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

// LDAPUserProvider is a provider using a LDAP or AD as a user database.
type LDAPUserProvider struct {
	configuration     schema.LDAPAuthenticationBackendConfiguration
	tlsConfig         *tls.Config
	dialOpts          ldap.DialOpt
	logger            *logrus.Logger
	connectionFactory LDAPConnectionFactory
	usersBaseDN       string
	groupsBaseDN      string
	profileAttributes []string

	supportExtensionPasswdModify bool
}

// NewLDAPUserProvider creates a new instance of LDAPUserProvider.
func NewLDAPUserProvider(configuration schema.AuthenticationBackendConfiguration, certPool *x509.CertPool) (provider *LDAPUserProvider, err error) {
	provider = newLDAPUserProvider(*configuration.LDAP, certPool, nil)

	err = provider.checkServer()
	if err != nil {
		return provider, err
	}

	if !provider.supportExtensionPasswdModify && !configuration.DisableResetPassword &&
		provider.configuration.Implementation != schema.LDAPImplementationActiveDirectory {
		provider.logger.Warnf("Your LDAP server implementation may not support a method for password hashing " +
			"known to Authelia, it's strongly recommended you ensure your directory server hashes the password " +
			"attribute when users reset their password via Authelia.")
	}

	return provider, nil
}

func newLDAPUserProvider(configuration schema.LDAPAuthenticationBackendConfiguration, certPool *x509.CertPool, factory LDAPConnectionFactory) (provider *LDAPUserProvider) {
	if configuration.TLS == nil {
		configuration.TLS = schema.DefaultLDAPAuthenticationBackendConfiguration.TLS
	}

	tlsConfig := utils.NewTLSConfig(configuration.TLS, tls.VersionTLS12, certPool)

	var dialOpts ldap.DialOpt

	if tlsConfig != nil {
		dialOpts = ldap.DialWithTLSConfig(tlsConfig)
	}

	if factory == nil {
		factory = NewLDAPConnectionFactoryImpl()
	}

	provider = &LDAPUserProvider{
		configuration:     configuration,
		tlsConfig:         tlsConfig,
		dialOpts:          dialOpts,
		logger:            logging.Logger(),
		connectionFactory: factory,
	}

	provider.parseDynamicConfiguration()

	return provider
}

func (p *LDAPUserProvider) parseDynamicConfiguration() {
	p.profileAttributes = []string{
		p.configuration.DisplayNameAttribute,
		p.configuration.MailAttribute,
		p.configuration.UsernameAttribute,
	}

	if p.configuration.GroupsAttribute != "" {
		p.profileAttributes = append(p.profileAttributes, p.configuration.GroupsAttribute)
	}

	p.configuration.UsersFilter = strings.ReplaceAll(p.configuration.UsersFilter, "{username_attribute}", p.configuration.UsernameAttribute)
	p.configuration.UsersFilter = strings.ReplaceAll(p.configuration.UsersFilter, "{mail_attribute}", p.configuration.MailAttribute)
	p.configuration.UsersFilter = strings.ReplaceAll(p.configuration.UsersFilter, "{display_name_attribute}", p.configuration.DisplayNameAttribute)

	p.logger.Tracef("Dynamically generated users filter is %s", p.configuration.UsersFilter)

	if p.configuration.AdditionalUsersDN != "" {
		p.usersBaseDN = p.configuration.AdditionalUsersDN + "," + p.configuration.BaseDN
	} else {
		p.usersBaseDN = p.configuration.BaseDN
	}

	p.logger.Tracef("Dynamically generated users BaseDN is %s", p.usersBaseDN)

	if p.configuration.AdditionalGroupsDN != "" {
		p.groupsBaseDN = ldap.EscapeFilter(p.configuration.AdditionalGroupsDN + "," + p.configuration.BaseDN)
	} else {
		p.groupsBaseDN = p.configuration.BaseDN
	}

	p.logger.Tracef("Dynamically generated groups BaseDN is %s", p.groupsBaseDN)
}

func (p *LDAPUserProvider) checkServer() (err error) {
	conn, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return err
	}

	searchRequest := ldap.NewSearchRequest("", ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		1, 0, false, "(objectClass=*)", []string{ldapSupportedExtensionAttribute}, nil)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return err
	}

	if len(sr.Entries) != 1 {
		return nil
	}

	// Iterate the attribute values to see what the server supports.
	for _, attr := range sr.Entries[0].Attributes {
		if attr.Name == ldapSupportedExtensionAttribute {
			p.logger.Tracef("LDAP Supported Extension OIDs: %s", strings.Join(attr.Values, ", "))

			for _, oid := range attr.Values {
				if oid == ldapOIDPasswdModifyExtension {
					p.supportExtensionPasswdModify = true
					break
				}
			}

			break
		}
	}

	return nil
}

func (p *LDAPUserProvider) connect(userDN string, password string) (LDAPConnection, error) {
	conn, err := p.connectionFactory.DialURL(p.configuration.URL, p.dialOpts)
	if err != nil {
		return nil, err
	}

	if p.configuration.StartTLS {
		if err := conn.StartTLS(p.tlsConfig); err != nil {
			return nil, err
		}
	}

	if err := conn.Bind(userDN, password); err != nil {
		return nil, err
	}

	return conn, nil
}

// CheckUserPassword checks if provided password matches for the given user.
func (p *LDAPUserProvider) CheckUserPassword(inputUsername string, password string) (bool, error) {
	conn, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	profile, err := p.getUserProfile(conn, inputUsername)
	if err != nil {
		return false, err
	}

	userConn, err := p.connect(profile.DN, password)
	if err != nil {
		return false, fmt.Errorf("Authentication of user %s failed. Cause: %s", inputUsername, err)
	}
	defer userConn.Close()

	return true, nil
}

func (p *LDAPUserProvider) ldapEscape(inputUsername string) string {
	inputUsername = ldap.EscapeFilter(inputUsername)
	for _, c := range specialLDAPRunes {
		inputUsername = strings.ReplaceAll(inputUsername, string(c), fmt.Sprintf("\\%c", c))
	}

	return inputUsername
}

type ldapUserProfile struct {
	DN          string
	Emails      []string
	DisplayName string
	Username    string
	Groups      []string
}

func (p *LDAPUserProvider) resolveUsersFilter(userFilter string, inputUsername string) string {
	inputUsername = p.ldapEscape(inputUsername)

	// The {input} placeholder is replaced by the users username input.
	userFilter = strings.ReplaceAll(userFilter, "{input}", inputUsername)

	p.logger.Tracef("Computed user filter is %s", userFilter)

	return userFilter
}

func (p *LDAPUserProvider) getUserProfile(conn LDAPConnection, inputUsername string) (*ldapUserProfile, error) {
	userFilter := p.resolveUsersFilter(p.configuration.UsersFilter, inputUsername)

	// Search for the given username.
	searchRequest := ldap.NewSearchRequest(
		p.usersBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		1, 0, false, userFilter, p.profileAttributes, nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("Cannot find user DN of user %s. Cause: %s", inputUsername, err)
	}

	if len(sr.Entries) == 0 {
		return nil, ErrUserNotFound
	}

	if len(sr.Entries) > 1 {
		return nil, fmt.Errorf("Multiple users %s found", inputUsername)
	}

	userProfile := ldapUserProfile{
		DN: sr.Entries[0].DN,
	}

	for _, attr := range sr.Entries[0].Attributes {
		if attr.Name == p.configuration.DisplayNameAttribute {
			userProfile.DisplayName = attr.Values[0]
		}

		if attr.Name == p.configuration.MailAttribute {
			userProfile.Emails = attr.Values
		}

		if attr.Name == p.configuration.UsernameAttribute {
			if len(attr.Values) != 1 {
				return nil, fmt.Errorf("User %s cannot have multiple value for attribute %s",
					inputUsername, p.configuration.UsernameAttribute)
			}

			userProfile.Username = attr.Values[0]
		}

		if attr.Name == p.configuration.GroupsAttribute && p.configuration.GroupsAttribute != "" {
			userProfile.Groups = attr.Values
		}
	}

	if userProfile.DN == "" {
		return nil, fmt.Errorf("No DN has been found for user %s", inputUsername)
	}

	return &userProfile, nil
}

func (p *LDAPUserProvider) resolveGroupsFilter(inputUsername string, profile *ldapUserProfile) (filter string, err error) {
	if p.configuration.GroupsAttribute != "" {
		filter = ldapBuildGroupsFilterFromGroupsAttribute(profile.Groups, p.configuration.DistinguishedNameAttribute)

		if filter == "" {
			return filter, errEmptyGroupsFilter
		}

		return filter, nil
	}

	inputUsername = p.ldapEscape(inputUsername)

	// The {input} placeholder is replaced by the users username input.
	groupFilter := strings.ReplaceAll(p.configuration.GroupsFilter, "{input}", inputUsername)

	if profile != nil {
		groupFilter = strings.ReplaceAll(groupFilter, "{username}", ldap.EscapeFilter(profile.Username))
		groupFilter = strings.ReplaceAll(groupFilter, "{dn}", ldap.EscapeFilter(profile.DN))
	}

	p.logger.Tracef("Computed groups filter is %s", groupFilter)

	return groupFilter, nil
}

// GetDetails retrieve the groups a user belongs to.
func (p *LDAPUserProvider) GetDetails(inputUsername string) (*UserDetails, error) {
	conn, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	profile, err := p.getUserProfile(conn, inputUsername)
	if err != nil {
		return nil, err
	}

	groupsFilter, err := p.resolveGroupsFilter(inputUsername, profile)

	var groups []string

	if err != errEmptyGroupsFilter {
		// Search for the given username.
		searchGroupRequest := ldap.NewSearchRequest(
			p.groupsBaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
			0, 0, false, groupsFilter, []string{p.configuration.GroupNameAttribute}, nil,
		)

		sr, err := conn.Search(searchGroupRequest)

		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve groups of user %s. Cause: %s", inputUsername, err)
		}

		for _, res := range sr.Entries {
			if len(res.Attributes) == 0 {
				p.logger.Warningf("No groups retrieved from LDAP for user %s", inputUsername)
				break
			}
			// Append all values of the document. Normally there should be only one per document.
			groups = append(groups, res.Attributes[0].Values...)
		}
	}

	return &UserDetails{
		Username:    profile.Username,
		DisplayName: profile.DisplayName,
		Emails:      profile.Emails,
		Groups:      groups,
	}, nil
}

// UpdatePassword update the password of the given user.
func (p *LDAPUserProvider) UpdatePassword(inputUsername string, newPassword string) error {
	conn, err := p.connect(p.configuration.User, p.configuration.Password)
	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}
	defer conn.Close()

	profile, err := p.getUserProfile(conn, inputUsername)

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	switch {
	case p.supportExtensionPasswdModify:
		modifyRequest := ldap.NewPasswordModifyRequest(
			profile.DN,
			"",
			newPassword,
		)

		err = conn.PasswordModify(modifyRequest)
	case p.configuration.Implementation == schema.LDAPImplementationActiveDirectory:
		modifyRequest := ldap.NewModifyRequest(profile.DN, nil)
		utf16 := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
		// The password needs to be enclosed in quotes
		// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/6e803168-f140-4d23-b2d3-c3a8ab5917d2
		pwdEncoded, _ := utf16.NewEncoder().String(fmt.Sprintf("\"%s\"", newPassword))
		modifyRequest.Replace("unicodePwd", []string{pwdEncoded})

		err = conn.Modify(modifyRequest)
	default:
		modifyRequest := ldap.NewModifyRequest(profile.DN, nil)
		modifyRequest.Replace("userPassword", []string{newPassword})

		err = conn.Modify(modifyRequest)
	}

	if err != nil {
		return fmt.Errorf("Unable to update password. Cause: %s", err)
	}

	return nil
}
