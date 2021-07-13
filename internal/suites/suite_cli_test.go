package suites

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CLISuite struct {
	*CommandSuite
}

func NewCLISuite() *CLISuite {
	return &CLISuite{CommandSuite: new(CommandSuite)}
}

func (s *CLISuite) SetupSuite() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/docker-compose.yml",
		"internal/suites/CLI/docker-compose.yml",
		"internal/suites/example/compose/authelia/docker-compose.backend.{}.yml",
	})
	s.DockerEnvironment = dockerEnvironment
}

func (s *CLISuite) SetupTest() {
	testArg := ""
	coverageArg := ""

	if os.Getenv("CI") == stringTrue {
		testArg = "-test.coverprofile=/authelia/coverage-$(date +%s).txt"
		coverageArg = "COVERAGE"
	}

	s.testArg = testArg
	s.coverageArg = coverageArg
}

func (s *CLISuite) TestShouldPrintBuildInformation() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "build-info"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Last Tag: ")
	s.Assert().Contains(output, "State: ")
	s.Assert().Contains(output, "Branch: ")
	s.Assert().Contains(output, "Build Number: ")
	s.Assert().Contains(output, "Build OS: ")
	s.Assert().Contains(output, "Build Arch: ")
	s.Assert().Contains(output, "Build Date: ")

	r := regexp.MustCompile(`^Last Tag: v\d+\.\d+\.\d+\nState: (tagged|untagged) (clean|dirty)\nBranch: [^\s\n]+\nCommit: [0-9a-f]{40}\nBuild Number: \d+\nBuild OS: (linux|darwin|windows|freebsd)\nBuild Arch: (amd64|arm|arm64)\nBuild Date: (Sun|Mon|Tue|Wed|Thu|Fri|Sat), \d{2} (Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d{4} \d{2}:\d{2}:\d{2} [+-]\d{4}\nExtra: \n`)
	s.Assert().Regexp(r, output)
}

func (s *CLISuite) TestShouldPrintVersion() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "--version"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "authelia version")
}

func (s *CLISuite) TestShouldValidateConfig() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "validate-config", "/config/configuration.yml"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Configuration parsed successfully without errors")
}

func (s *CLISuite) TestShouldFailValidateConfig() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "validate-config", "/config/invalid.yml"})
	s.Assert().NotNil(err)
	s.Assert().Contains(output, "Error Loading Configuration: stat /config/invalid.yml: no such file or directory")
}

func (s *CLISuite) TestShouldHashPasswordArgon2id() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "hash-password", "test", "-m", "32", "-s", "test1234"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Password hash: $argon2id$v=19$m=32768,t=1,p=8")
}

func (s *CLISuite) TestShouldHashPasswordSHA512() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "hash-password", "test", "-z"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Password hash: $6$rounds=50000")
}

func (s *CLISuite) TestShouldGenerateCertificateRSA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateRSAWithIPAddress() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "127.0.0.1", "--dir", "/tmp/"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateRSAWithStartDate() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--start-date", "'Jan 1 15:04:05 2011'"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldFailGenerateCertificateRSAWithStartDate() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--start-date", "Jan"})
	s.Assert().NotNil(err)
	s.Assert().Contains(output, "Failed to parse start date: parsing time \"Jan\" as \"Jan 2 15:04:05 2006\": cannot parse \"\" as \"2\"")
}

func (s *CLISuite) TestShouldGenerateCertificateCA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ca"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateEd25519() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ed25519"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldFailGenerateCertificateECDSA() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ecdsa-curve", "invalid"})
	s.Assert().NotNil(err)
	s.Assert().Contains(output, "Failed to generate private key: unrecognized elliptic curve: \"invalid\"")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSAP224() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ecdsa-curve", "P224"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSAP256() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ecdsa-curve", "P256"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSAP384() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ecdsa-curve", "P384"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func (s *CLISuite) TestShouldGenerateCertificateECDSAP521() {
	output, err := s.Exec("authelia-backend", []string{"authelia", s.testArg, s.coverageArg, "certificates", "generate", "--host", "*.example.com", "--dir", "/tmp/", "--ecdsa-curve", "P521"})
	s.Assert().Nil(err)
	s.Assert().Contains(output, "Certificate Public Key written to /tmp/cert.pem")
	s.Assert().Contains(output, "Certificate Private Key written to /tmp/key.pem")
}

func TestCLISuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewCLISuite())
}
