package authentication

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

func TestShouldCreateRawConnectionWhenSchemeIsLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL: "ldap://127.0.0.1:389",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	_, err := ldapClient.connect("cn=admin,dc=example,dc=com", "password")

	require.NoError(t, err)
}

func TestShouldCreateTLSConnectionWhenSchemeIsLDAPS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL: "ldaps://127.0.0.1:389",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldaps://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	_, err := ldapClient.connect("cn=admin,dc=example,dc=com", "password")

	require.NoError(t, err)
}

func TestEscapeSpecialCharsFromUserInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL: "ldaps://127.0.0.1:389",
		},
		nil,
		mockFactory)

	// No escape
	assert.Equal(t, "xyz", ldapClient.ldapEscape("xyz"))

	// Escape
	assert.Equal(t, "test\\,abc", ldapClient.ldapEscape("test,abc"))
	assert.Equal(t, "test\\5cabc", ldapClient.ldapEscape("test\\abc"))
	assert.Equal(t, "test\\2aabc", ldapClient.ldapEscape("test*abc"))
	assert.Equal(t, "test \\28abc\\29", ldapClient.ldapEscape("test (abc)"))
	assert.Equal(t, "test\\#abc", ldapClient.ldapEscape("test#abc"))
	assert.Equal(t, "test\\+abc", ldapClient.ldapEscape("test+abc"))
	assert.Equal(t, "test\\<abc", ldapClient.ldapEscape("test<abc"))
	assert.Equal(t, "test\\>abc", ldapClient.ldapEscape("test>abc"))
	assert.Equal(t, "test\\;abc", ldapClient.ldapEscape("test;abc"))
	assert.Equal(t, "test\\\"abc", ldapClient.ldapEscape("test\"abc"))
	assert.Equal(t, "test\\=abc", ldapClient.ldapEscape("test=abc"))
	assert.Equal(t, "test\\,\\5c\\28abc\\29", ldapClient.ldapEscape("test,\\(abc)"))
}

func TestEscapeSpecialCharsInGroupsFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:          "ldaps://127.0.0.1:389",
			GroupsFilter: "(|(member={dn})(uid={username})(uid={input}))",
		},
		nil,
		mockFactory)

	profile := ldapUserProfile{
		DN:          "cn=john (external),dc=example,dc=com",
		Username:    "john",
		DisplayName: "John Doe",
		Emails:      []string{"john.doe@authelia.com"},
	}

	filter, err := ldapClient.resolveGroupsFilter("john", &profile)
	assert.NoError(t, err)
	assert.Equal(t, "(|(member=cn=john \\28external\\29,dc=example,dc=com)(uid=john)(uid=john))", filter)

	filter, _ = ldapClient.resolveGroupsFilter("john#=(abc,def)", &profile)
	assert.Equal(t, "(|(member=cn=john \\28external\\29,dc=example,dc=com)(uid=john)(uid=john\\#\\=\\28abc\\,def\\29))", filter)
}

type ExtendedSearchRequestMatcher struct {
	filter       string
	baseDN       string
	scope        int
	derefAliases int
	typesOnly    bool
	attributes   []string
}

func NewExtendedSearchRequestMatcher(filter, base string, scope, derefAliases int, typesOnly bool, attributes []string) *ExtendedSearchRequestMatcher {
	return &ExtendedSearchRequestMatcher{filter, base, scope, derefAliases, typesOnly, attributes}
}

func (e *ExtendedSearchRequestMatcher) Matches(x interface{}) bool {
	sr := x.(*ldap.SearchRequest)

	if e.filter != sr.Filter || e.baseDN != sr.BaseDN || e.scope != sr.Scope || e.derefAliases != sr.DerefAliases ||
		e.typesOnly != sr.TypesOnly || utils.IsStringSlicesDifferent(e.attributes, sr.Attributes) {
		return false
	}

	return true
}

func (e *ExtendedSearchRequestMatcher) String() string {
	return fmt.Sprintf("baseDN: %s, filter %s", e.baseDN, e.filter)
}

func TestShouldCheckLDAPServerExtensions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{ldapOIDPasswdModifyExtension},
						},
					},
				},
			},
		}, nil)

	err := ldapClient.checkServer()
	assert.NoError(t, err)

	assert.True(t, ldapClient.supportExtensionPasswdModify)
}

func TestShouldNotEnablePasswdModifyExtension(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute})).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   ldapSupportedExtensionAttribute,
							Values: []string{},
						},
					},
				},
			},
		}, nil)

	err := ldapClient.checkServer()
	assert.NoError(t, err)

	assert.False(t, ldapClient.supportExtensionPasswdModify)
}

func TestShouldReturnCheckServerConnectError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, errors.New("could not connect"))

	err := ldapClient.checkServer()
	assert.EqualError(t, err, "could not connect")

	assert.False(t, ldapClient.supportExtensionPasswdModify)
}

func TestShouldReturnCheckServerSearchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute})).
		Return(nil, errors.New("could not perform the search"))

	err := ldapClient.checkServer()
	assert.EqualError(t, err, "could not perform the search")

	assert.False(t, ldapClient.supportExtensionPasswdModify)
}

type SearchRequestMatcher struct {
	expected string
}

func NewSearchRequestMatcher(expected string) *SearchRequestMatcher {
	return &SearchRequestMatcher{expected}
}

func (srm *SearchRequestMatcher) Matches(x interface{}) bool {
	sr := x.(*ldap.SearchRequest)
	return sr.Filter == srm.expected
}

func (srm *SearchRequestMatcher) String() string {
	return ""
}

func TestShouldEscapeUserInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			UsersFilter:          "(|({username_attribute}={input})({mail_attribute}={input}))",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockConn.EXPECT().
		// Here we ensure that the input has been correctly escaped.
		Search(NewSearchRequestMatcher("(|(uid=john\\=abc)(mail=john\\=abc))")).
		Return(&ldap.SearchResult{}, nil)

	_, err := ldapClient.getUserProfile(mockConn, "john=abc")
	require.Error(t, err)
	assert.EqualError(t, err, "user not found")
}

func TestShouldCombineUsernameFilterAndUsersFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			UsernameAttribute:    "uid",
			UsersFilter:          "(&({username_attribute}={input})(&(objectCategory=person)(objectClass=user)))",
			Password:             "password",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
		},
		nil,
		mockFactory)

	mockConn.EXPECT().
		Search(NewSearchRequestMatcher("(&(uid=john)(&(objectCategory=person)(objectClass=user)))")).
		Return(&ldap.SearchResult{}, nil)

	_, err := ldapClient.getUserProfile(mockConn, "john")
	require.Error(t, err)
	assert.EqualError(t, err, "user not found")
}

func createSearchResultWithAttributes(attributes ...*ldap.EntryAttribute) *ldap.SearchResult {
	return &ldap.SearchResult{
		Entries: []*ldap.Entry{
			{Attributes: attributes},
		},
	}
}

func createSearchResultWithAttributeValues(values ...string) *ldap.SearchResult {
	return createSearchResultWithAttributes(&ldap.EntryAttribute{
		Values: values,
	})
}

func TestShouldNotCrashWhenGroupsAreNotRetrievedFromLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		Close()

	searchGroups := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributes(), nil)
	searchProfile := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayname",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(searchProfile, searchGroups)

	details, err := ldapClient.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "john")
}

func TestShouldNotCrashWhenEmailsAreNotRetrievedFromLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:               "ldap://127.0.0.1:389",
			User:              "cn=admin,dc=example,dc=com",
			Password:          "password",
			UsernameAttribute: "uid",
			UsersFilter:       "uid={input}",
			AdditionalUsersDN: "ou=users",
			BaseDN:            "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		Close()

	searchGroups := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributeValues("group1", "group2"), nil)
	searchProfile := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "uid",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(searchProfile, searchGroups)

	details, err := ldapClient.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{})
	assert.Equal(t, details.Username, "john")
}

func TestShouldReturnUsernameFromLDAP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		Close()

	searchGroups := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributeValues("group1", "group2"), nil)
	searchProfile := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayname",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"John"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(searchProfile, searchGroups)

	details, err := ldapClient.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{"group1", "group2"})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "John")
}

func TestShouldUpdateUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	pwdModifyRequest := ldap.NewPasswordModifyRequest(
		"uid=test,dc=example,dc=com",
		"",
		"password",
	)

	gomock.InOrder(
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockConn, nil),
		mockConn.EXPECT().
			Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),

		mockConn.EXPECT().
			Search(NewExtendedSearchRequestMatcher("(objectClass=*)", "", ldap.ScopeBaseObject, ldap.NeverDerefAliases, false, []string{ldapSupportedExtensionAttribute})).
			Return(&ldap.SearchResult{
				Entries: []*ldap.Entry{
					{
						DN: "",
						Attributes: []*ldap.EntryAttribute{
							{
								Name:   ldapSupportedExtensionAttribute,
								Values: []string{ldapOIDPasswdModifyExtension},
							},
						},
					},
				},
			}, nil),

		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockConn, nil),
		mockConn.EXPECT().
			Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockConn.EXPECT().
			Search(gomock.Any()).
			Return(&ldap.SearchResult{
				Entries: []*ldap.Entry{
					{
						DN: "uid=test,dc=example,dc=com",
						Attributes: []*ldap.EntryAttribute{
							{
								Name:   "displayname",
								Values: []string{"John Doe"},
							},
							{
								Name:   "mail",
								Values: []string{"test@example.com"},
							},
							{
								Name:   "uid",
								Values: []string{"John"},
							},
						},
					},
				},
			}, nil),
		mockConn.EXPECT().
			PasswordModify(pwdModifyRequest).
			Return(nil),
		mockConn.EXPECT().
			Close(),
	)

	err := ldapClient.checkServer()
	require.NoError(t, err)

	err = ldapClient.UpdatePassword("john", "password")
	require.NoError(t, err)
}

func TestShouldCheckValidUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	gomock.InOrder(
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockConn, nil),
		mockConn.EXPECT().
			Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockConn.EXPECT().
			Search(gomock.Any()).
			Return(&ldap.SearchResult{
				Entries: []*ldap.Entry{
					{
						DN: "uid=test,dc=example,dc=com",
						Attributes: []*ldap.EntryAttribute{
							{
								Name:   "displayname",
								Values: []string{"John Doe"},
							},
							{
								Name:   "mail",
								Values: []string{"test@example.com"},
							},
							{
								Name:   "uid",
								Values: []string{"John"},
							},
						},
					},
				},
			}, nil),
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockConn, nil),
		mockConn.EXPECT().
			Bind(gomock.Eq("uid=test,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockConn.EXPECT().
			Close().Times(2),
	)

	valid, err := ldapClient.CheckUserPassword("john", "password")

	assert.True(t, valid)
	require.NoError(t, err)
}

func TestShouldCheckInvalidUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
		},
		nil,
		mockFactory)

	gomock.InOrder(
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockConn, nil),
		mockConn.EXPECT().
			Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
			Return(nil),
		mockConn.EXPECT().
			Search(gomock.Any()).
			Return(&ldap.SearchResult{
				Entries: []*ldap.Entry{
					{
						DN: "uid=test,dc=example,dc=com",
						Attributes: []*ldap.EntryAttribute{
							{
								Name:   "displayname",
								Values: []string{"John Doe"},
							},
							{
								Name:   "mail",
								Values: []string{"test@example.com"},
							},
							{
								Name:   "uid",
								Values: []string{"John"},
							},
						},
					},
				},
			}, nil),
		mockFactory.EXPECT().
			DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
			Return(mockConn, nil),
		mockConn.EXPECT().
			Bind(gomock.Eq("uid=test,dc=example,dc=com"), gomock.Eq("password")).
			Return(errors.New("Invalid username or password")),
		mockConn.EXPECT().
			Close(),
	)

	valid, err := ldapClient.CheckUserPassword("john", "password")

	assert.False(t, valid)
	require.EqualError(t, err, "Authentication of user john failed. Cause: Invalid username or password")
}

func TestShouldCallStartTLSWhenEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		StartTLS(ldapClient.tlsConfig)

	mockConn.EXPECT().
		Close()

	searchGroups := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributes(), nil)
	searchProfile := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayname",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(searchProfile, searchGroups)

	details, err := ldapClient.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "john")
}

func TestShouldParseDynamicConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "(&(|({username_attribute}={input})({mail_attribute}={input})({display_name_attribute}={input}))(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2)(!pwdLastSet=0))",
			GroupsFilter:         "(&(|(member={dn})(member={input})(member={username}))(objectClass=group))",
			AdditionalUsersDN:    "ou=users",
			AdditionalGroupsDN:   "ou=groups",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
		},
		nil,
		mockFactory)

	assert.Equal(t, "(&(|(uid={input})(mail={input})(displayname={input}))(objectCategory=person)(objectClass=user)(!userAccountControl:1.2.840.113556.1.4.803:=2)(!pwdLastSet=0))", ldapClient.configuration.UsersFilter)
	assert.Equal(t, "(&(|(member={dn})(member={input})(member={username}))(objectClass=group))", ldapClient.configuration.GroupsFilter)
	assert.Equal(t, "ou=users,dc=example,dc=com", ldapClient.usersBaseDN)
	assert.Equal(t, "ou=groups,dc=example,dc=com", ldapClient.groupsBaseDN)
}

func TestShouldCallStartTLSWithInsecureSkipVerifyWhenSkipVerifyTrue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldap://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
			TLS: &schema.TLSConfig{
				SkipVerify: true,
			},
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldap://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		Bind(gomock.Eq("cn=admin,dc=example,dc=com"), gomock.Eq("password")).
		Return(nil)

	mockConn.EXPECT().
		StartTLS(ldapClient.tlsConfig)

	mockConn.EXPECT().
		Close()

	searchGroups := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(createSearchResultWithAttributes(), nil)
	searchProfile := mockConn.EXPECT().
		Search(gomock.Any()).
		Return(&ldap.SearchResult{
			Entries: []*ldap.Entry{
				{
					DN: "uid=test,dc=example,dc=com",
					Attributes: []*ldap.EntryAttribute{
						{
							Name:   "displayname",
							Values: []string{"John Doe"},
						},
						{
							Name:   "mail",
							Values: []string{"test@example.com"},
						},
						{
							Name:   "uid",
							Values: []string{"john"},
						},
					},
				},
			},
		}, nil)

	gomock.InOrder(searchProfile, searchGroups)

	details, err := ldapClient.GetDetails("john")
	require.NoError(t, err)

	assert.ElementsMatch(t, details.Groups, []string{})
	assert.ElementsMatch(t, details.Emails, []string{"test@example.com"})
	assert.Equal(t, details.DisplayName, "John Doe")
	assert.Equal(t, details.Username, "john")
}

func TestShouldReturnLDAPSAlreadySecuredWhenStartTLSAttempted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFactory := NewMockLDAPConnectionFactory(ctrl)
	mockConn := NewMockLDAPConnection(ctrl)

	ldapClient := newLDAPUserProvider(
		schema.LDAPAuthenticationBackendConfiguration{
			URL:                  "ldaps://127.0.0.1:389",
			User:                 "cn=admin,dc=example,dc=com",
			Password:             "password",
			UsernameAttribute:    "uid",
			MailAttribute:        "mail",
			DisplayNameAttribute: "displayname",
			UsersFilter:          "uid={input}",
			AdditionalUsersDN:    "ou=users",
			BaseDN:               "dc=example,dc=com",
			StartTLS:             true,
			TLS: &schema.TLSConfig{
				SkipVerify: true,
			},
		},
		nil,
		mockFactory)

	mockFactory.EXPECT().
		DialURL(gomock.Eq("ldaps://127.0.0.1:389"), gomock.Any()).
		Return(mockConn, nil)

	mockConn.EXPECT().
		StartTLS(ldapClient.tlsConfig).
		Return(errors.New("LDAP Result Code 200 \"Network Error\": ldap: already encrypted"))

	_, err := ldapClient.GetDetails("john")
	assert.EqualError(t, err, "LDAP Result Code 200 \"Network Error\": ldap: already encrypted")
}
