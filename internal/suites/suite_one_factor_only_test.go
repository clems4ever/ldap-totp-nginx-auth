package suites

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type OneFactorOnlySuite struct {
	suite.Suite
}

type OneFactorOnlyWebSuite struct {
	*SeleniumSuite
}

func NewOneFactorOnlyWebSuite() *OneFactorOnlyWebSuite {
	return &OneFactorOnlyWebSuite{SeleniumSuite: new(SeleniumSuite)}
}

func (s *OneFactorOnlyWebSuite) SetupSuite() {
	wds, err := StartWebDriver()

	if err != nil {
		log.Fatal(err)
	}

	s.WebDriverSession = wds
}

func (s *OneFactorOnlyWebSuite) TearDownSuite() {
	err := s.WebDriverSession.Stop()

	if err != nil {
		log.Fatal(err)
	}
}

func (s *OneFactorOnlyWebSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLogout(ctx, s.T())
}

// No target url is provided, then the user should be redirect to the default url.
func (s *OneFactorOnlyWebSuite) TestShouldRedirectUserToDefaultURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")
}

// Unsafe URL is provided, then the user should be redirect to the default url.
func (s *OneFactorOnlyWebSuite) TestShouldRedirectUserToDefaultURLWhenURLIsUnsafe() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "http://unsafe.local")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")
}

// When use logged in and visit the portal again, she gets redirect to the authenticated view.
func (s *OneFactorOnlyWebSuite) TestShouldDisplayAuthenticatedView() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")
	s.doVisit(s.T(), GetLoginBaseURL())
	s.verifyIsAuthenticatedPage(ctx, s.T())
}

func (s *OneFactorOnlyWebSuite) TestShouldRedirectAlreadyAuthenticatedUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")

	s.doVisit(s.T(), fmt.Sprintf("%s?rd=https://singlefactor.example.com:8080/secret.html", GetLoginBaseURL()))
	s.verifyURLIs(ctx, s.T(), "https://singlefactor.example.com:8080/secret.html")
}

func (s *OneFactorOnlyWebSuite) TestShouldNotRedirectAlreadyAuthenticatedUserToUnsafeURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	s.doLoginOneFactor(ctx, s.T(), "john", "password", false, "")
	s.verifyURLIs(ctx, s.T(), HomeBaseURL+"/")

	// Visit the login page and wait for redirection to 2FA page with success icon displayed.
	s.doVisit(s.T(), fmt.Sprintf("%s?rd=https://secure.example.local:8080", GetLoginBaseURL()))
	s.verifyNotificationDisplayed(ctx, s.T(), "There was an issue redirecting the user. Check that the redirection URI matches the domain.")
}

func (s *OneFactorOnlySuite) TestWeb() {
	suite.Run(s.T(), NewOneFactorOnlyWebSuite())
}

func TestOneFactorOnlySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, new(OneFactorOnlySuite))
}
