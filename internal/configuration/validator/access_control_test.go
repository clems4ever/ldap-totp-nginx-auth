package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/configuration/schema"
)

type AccessControl struct {
	suite.Suite
	configuration schema.AccessControlConfiguration
	validator     *schema.StructValidator
}

func (suite *AccessControl) SetupTest() {
	suite.validator = schema.NewStructValidator()
	suite.configuration.DefaultPolicy = policyDeny
	suite.configuration.Networks = schema.DefaultACLNetwork
	suite.configuration.Rules = schema.DefaultACLRule
}

func (suite *AccessControl) TestShouldValidateCompleteConfiguration() {
	ValidateAccessControl(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Assert().False(suite.validator.HasErrors())
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidDefaultPolicy() {
	suite.configuration.DefaultPolicy = testInvalidPolicy

	ValidateAccessControl(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "'default_policy' must either be 'deny', 'two_factor', 'one_factor' or 'bypass'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidNetworkGroupNetwork() {
	suite.configuration.Networks = []schema.ACLNetwork{
		{
			Name:     "internal",
			Networks: []string{"abc.def.ghi.jkl"},
		},
	}

	ValidateAccessControl(&suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Network [abc.def.ghi.jkl] from network group: internal must be a valid IP or CIDR")
}

func (suite *AccessControl) TestShouldRaiseErrorWithNoRulesDefined() {
	suite.configuration.Rules = []schema.ACLRule{}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Default Policy [deny] is invalid, access control rules must be provided or a policy must either be 'one_factor' or 'two_factor'")
}

func (suite *AccessControl) TestShouldRaiseWarningWithNoRulesDefined() {
	suite.configuration.Rules = []schema.ACLRule{}

	suite.configuration.DefaultPolicy = policyTwoFactor

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasErrors())
	suite.Require().Len(suite.validator.Warnings(), 1)

	suite.Assert().EqualError(suite.validator.Warnings()[0], "No access control rules have been defined so the default policy two_factor will be applied to all requests")
}

func (suite *AccessControl) TestShouldRaiseErrorsWithEmptyRules() {
	suite.configuration.Rules = []schema.ACLRule{{}, {}}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 4)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Rule #1 is invalid, a policy must have one or more domains")
	suite.Assert().EqualError(suite.validator.Errors()[1], "Policy [] for rule #1 domain: [] is invalid, a policy must either be 'deny', 'two_factor', 'one_factor' or 'bypass'")
	suite.Assert().EqualError(suite.validator.Errors()[2], "Rule #2 is invalid, a policy must have one or more domains")
	suite.Assert().EqualError(suite.validator.Errors()[3], "Policy [] for rule #2 domain: [] is invalid, a policy must either be 'deny', 'two_factor', 'one_factor' or 'bypass'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidPolicy() {
	suite.configuration.Rules = []schema.ACLRule{
		{
			Domains: []string{"public.example.com"},
			Policy:  testInvalidPolicy,
		},
	}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Policy [invalid] for rule #1 domain: [public.example.com] is invalid, a policy must either be 'deny', 'two_factor', 'one_factor' or 'bypass'")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidNetwork() {
	suite.configuration.Rules = []schema.ACLRule{
		{
			Domains:  []string{"public.example.com"},
			Policy:   "bypass",
			Networks: []string{"abc.def.ghi.jkl/32"},
		},
	}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Network [abc.def.ghi.jkl/32] for rule #1 domain: [public.example.com] is not a valid network or network group")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidMethod() {
	suite.configuration.Rules = []schema.ACLRule{
		{
			Domains: []string{"public.example.com"},
			Policy:  "bypass",
			Methods: []string{"GET", "HOP"},
		},
	}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Method HOP for rule #1 domain: [public.example.com] is invalid, must be one of the following methods: GET, HEAD, POST, PUT, PATCH, DELETE, TRACE, CONNECT, OPTIONS")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidResource() {
	suite.configuration.Rules = []schema.ACLRule{
		{
			Domains:   []string{"public.example.com"},
			Policy:    "bypass",
			Resources: []string{"^/(api.*"},
		},
	}

	ValidateRules(suite.configuration, suite.validator)

	suite.Assert().False(suite.validator.HasWarnings())
	suite.Require().Len(suite.validator.Errors(), 1)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Resource [^/(api.*] for rule #1 domain: [public.example.com] is invalid, error parsing regexp: missing closing ): `^/(api.*`")
}

func (suite *AccessControl) TestShouldRaiseErrorInvalidSubject() {
	domains := []string{"public.example.com"}
	subjects := [][]string{{"invalid"}}
	suite.configuration.Rules = []schema.ACLRule{
		{
			Domains:  domains,
			Policy:   "bypass",
			Subjects: subjects,
		},
	}

	ValidateRules(suite.configuration, suite.validator)

	suite.Require().Len(suite.validator.Warnings(), 0)
	suite.Require().Len(suite.validator.Errors(), 2)

	suite.Assert().EqualError(suite.validator.Errors()[0], "Subject [invalid] for rule #1 domain: [public.example.com] is invalid, must start with 'user:' or 'group:'")
	suite.Assert().EqualError(suite.validator.Errors()[1], fmt.Sprintf(errAccessControlInvalidPolicyWithSubjects, 1, domains, subjects))
}

func TestAccessControl(t *testing.T) {
	suite.Run(t, new(AccessControl))
}

func TestShouldReturnCorrectResultsForValidNetworkGroups(t *testing.T) {
	config := schema.AccessControlConfiguration{
		Networks: schema.DefaultACLNetwork,
	}

	validNetwork := IsNetworkGroupValid(config, "internal")
	invalidNetwork := IsNetworkGroupValid(config, "127.0.0.1")

	assert.True(t, validNetwork)
	assert.False(t, invalidNetwork)
}
