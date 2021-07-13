package suites

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/storage"
)

type StandaloneWebDriverSuite struct {
	*SeleniumSuite
}

func NewStandaloneWebDriverSuite() *StandaloneWebDriverSuite {
	return &StandaloneWebDriverSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *StandaloneWebDriverSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *StandaloneWebDriverSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *StandaloneWebDriverSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
	s.WebDriverSession.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())
}

func (s *StandaloneWebDriverSuite) TestShouldLetUserKnowHeIsAlreadyAuthenticated() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_ = s.doRegisterAndLogin2FA(ctx, s.T(), "john", "password", false, "")

	// Visit home page to change context.
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())

	// Visit the login page and wait for redirection to 2FA page with success icon displayed.
	s.doVisit(s.T(), GetLoginBaseURL())
	s.verifyIsAuthenticatedPage(ctx, s.T())
}

func (s *StandaloneWebDriverSuite) TestShouldRedirectAlreadyAuthenticatedUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_ = s.doRegisterAndLogin2FA(ctx, s.T(), "john", "password", false, "")

	// Visit home page to change context.
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())

	// Visit the login page and wait for redirection to 2FA page with success icon displayed.
	s.doVisit(s.T(), fmt.Sprintf("%s?rd=https://secure.example.com:8080", GetLoginBaseURL()))
	s.verifyURLIs(ctx, s.T(), "https://secure.example.com:8080/")
}

func (s *StandaloneWebDriverSuite) TestShouldNotRedirectAlreadyAuthenticatedUserToUnsafeURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_ = s.doRegisterAndLogin2FA(ctx, s.T(), "john", "password", false, "")

	// Visit home page to change context.
	s.doVisit(s.T(), HomeBaseURL)
	s.verifyIsHome(ctx, s.T())

	// Visit the login page and wait for redirection to 2FA page with success icon displayed.
	s.doVisit(s.T(), fmt.Sprintf("%s?rd=https://secure.example.local:8080", GetLoginBaseURL()))
	s.verifyNotificationDisplayed(ctx, s.T(), "There was an issue redirecting the user. Check that the redirection URI matches the domain.")
}

func (s *StandaloneWebDriverSuite) TestShouldCheckUserIsAskedToRegisterDevice() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	username := "john"
	password := "password"

	// Clean up any TOTP secret already in DB.
	provider := storage.NewSQLiteProvider("/tmp/db.sqlite3")
	require.NoError(s.T(), provider.DeleteTOTPSecret(username))

	// Login one factor.
	s.doLoginOneFactor(ctx, s.T(), username, password, false, "")

	// Check the user is asked to register a new device.
	s.WaitElementLocatedByClassName(ctx, s.T(), "state-not-registered")

	// Then register the TOTP factor.
	s.doRegisterTOTP(ctx, s.T())
	// And logout.
	s.doLogout(ctx, s.T())

	// Login one factor again.
	s.doLoginOneFactor(ctx, s.T(), username, password, false, "")

	// now the user should be asked to perform 2FA
	s.WaitElementLocatedByClassName(ctx, s.T(), "state-method")
}

type StandaloneSuite struct {
	suite.Suite
}

func NewStandaloneSuite() *StandaloneSuite {
	return &StandaloneSuite{}
}

func (s *StandaloneSuite) TestShouldRespectMethodsACL() {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/verify?rd=%s", AutheliaBaseURL, GetLoginBaseURL()), nil)
	s.Assert().NoError(err)
	req.Header.Set("X-Forwarded-Method", "GET")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", fmt.Sprintf("secure.%s", BaseDomain))
	req.Header.Set("X-Forwarded-URI", "/")

	client := NewHTTPClient()
	res, err := client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(res.StatusCode, 302)
	body, err := ioutil.ReadAll(res.Body)
	s.Assert().NoError(err)

	urlEncodedAdminURL := url.QueryEscape(SecureBaseURL + "/")
	s.Assert().Equal(fmt.Sprintf("Found. Redirecting to %s?rd=%s&rm=GET", GetLoginBaseURL(), urlEncodedAdminURL), string(body))

	req.Header.Set("X-Forwarded-Method", "OPTIONS")

	res, err = client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(res.StatusCode, 200)
}

// Standard case using nginx.
func (s *StandaloneSuite) TestShouldVerifyAPIVerifyUnauthorize() {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/verify", AutheliaBaseURL), nil)
	s.Assert().NoError(err)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Original-URL", AdminBaseURL)

	client := NewHTTPClient()
	res, err := client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(res.StatusCode, 401)
	body, err := ioutil.ReadAll(res.Body)
	s.Assert().NoError(err)
	s.Assert().Equal(string(body), "Unauthorized")
}

// Standard case using Kubernetes.
func (s *StandaloneSuite) TestShouldVerifyAPIVerifyRedirectFromXOriginalURL() {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/verify?rd=%s", AutheliaBaseURL, GetLoginBaseURL()), nil)
	s.Assert().NoError(err)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Original-URL", AdminBaseURL)

	client := NewHTTPClient()
	res, err := client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(res.StatusCode, 302)
	body, err := ioutil.ReadAll(res.Body)
	s.Assert().NoError(err)

	urlEncodedAdminURL := url.QueryEscape(AdminBaseURL)
	s.Assert().Equal(fmt.Sprintf("Found. Redirecting to %s?rd=%s", GetLoginBaseURL(), urlEncodedAdminURL), string(body))
}

func (s *StandaloneSuite) TestShouldVerifyAPIVerifyRedirectFromXOriginalHostURI() {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/verify?rd=%s", AutheliaBaseURL, GetLoginBaseURL()), nil)
	s.Assert().NoError(err)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "secure.example.com:8080")
	req.Header.Set("X-Forwarded-URI", "/")

	client := NewHTTPClient()
	res, err := client.Do(req)
	s.Assert().NoError(err)
	s.Assert().Equal(res.StatusCode, 302)
	body, err := ioutil.ReadAll(res.Body)
	s.Assert().NoError(err)

	urlEncodedAdminURL := url.QueryEscape(SecureBaseURL + "/")
	s.Assert().Equal(fmt.Sprintf("Found. Redirecting to %s?rd=%s", GetLoginBaseURL(), urlEncodedAdminURL), string(body))
}

func (s *StandaloneSuite) TestStandaloneWebDriverScenario() {
	suite.Run(s.T(), NewStandaloneWebDriverSuite())
}

func (s *StandaloneSuite) TestOneFactorScenario() {
	suite.Run(s.T(), NewOneFactorScenario())
}

func (s *StandaloneSuite) TestTwoFactorScenario() {
	suite.Run(s.T(), NewTwoFactorScenario())
}

func (s *StandaloneSuite) TestBypassPolicyScenario() {
	suite.Run(s.T(), NewBypassPolicyScenario())
}

func (s *StandaloneSuite) TestBackendProtectionScenario() {
	suite.Run(s.T(), NewBackendProtectionScenario())
}

func (s *StandaloneSuite) TestResetPasswordScenario() {
	suite.Run(s.T(), NewResetPasswordScenario())
}

func (s *StandaloneSuite) TestAvailableMethodsScenario() {
	suite.Run(s.T(), NewAvailableMethodsScenario([]string{"TIME-BASED ONE-TIME PASSWORD"}))
}

func (s *StandaloneSuite) TestRedirectionURLScenario() {
	suite.Run(s.T(), NewRedirectionURLScenario())
}

func (s *StandaloneSuite) TestRedirectionCheckScenario() {
	suite.Run(s.T(), NewRedirectionCheckScenario())
}

func TestStandaloneSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewStandaloneSuite())
}
