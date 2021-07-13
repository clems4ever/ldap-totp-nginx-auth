package validator

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// IsPolicyValid check if policy is valid.
func IsPolicyValid(policy string) (isValid bool) {
	return policy == policyDeny || policy == policyOneFactor || policy == policyTwoFactor || policy == policyBypass
}

// IsResourceValid check if a resource is valid.
func IsResourceValid(resource string) (err error) {
	_, err = regexp.Compile(resource)
	return err
}

// IsSubjectValid check if a subject is valid.
func IsSubjectValid(subject string) (isValid bool) {
	return subject == "" || strings.HasPrefix(subject, "user:") || strings.HasPrefix(subject, "group:")
}

// IsNetworkGroupValid check if a network group is valid.
func IsNetworkGroupValid(configuration schema.AccessControlConfiguration, network string) bool {
	for _, networks := range configuration.Networks {
		if network != networks.Name {
			continue
		} else {
			return true
		}
	}

	return false
}

// IsNetworkValid check if a network is valid.
func IsNetworkValid(network string) (isValid bool) {
	if net.ParseIP(network) == nil {
		_, _, err := net.ParseCIDR(network)
		return err == nil
	}

	return true
}

// ValidateAccessControl validates access control configuration.
func ValidateAccessControl(configuration *schema.AccessControlConfiguration, validator *schema.StructValidator) {
	if configuration.DefaultPolicy == "" {
		configuration.DefaultPolicy = policyDeny
	}

	if !IsPolicyValid(configuration.DefaultPolicy) {
		validator.Push(fmt.Errorf("'default_policy' must either be 'deny', 'two_factor', 'one_factor' or 'bypass'"))
	}

	if configuration.Networks != nil {
		for _, n := range configuration.Networks {
			for _, networks := range n.Networks {
				if !IsNetworkValid(networks) {
					validator.Push(fmt.Errorf("Network %s from network group: %s must be a valid IP or CIDR", n.Networks, n.Name))
				}
			}
		}
	}
}

// ValidateRules validates an ACL Rule configuration.
func ValidateRules(configuration schema.AccessControlConfiguration, validator *schema.StructValidator) {
	if configuration.Rules == nil || len(configuration.Rules) == 0 {
		if configuration.DefaultPolicy != policyOneFactor && configuration.DefaultPolicy != policyTwoFactor {
			validator.Push(fmt.Errorf("Default Policy [%s] is invalid, access control rules must be provided or a policy must either be 'one_factor' or 'two_factor'", configuration.DefaultPolicy))

			return
		}

		validator.PushWarning(fmt.Errorf("No access control rules have been defined so the default policy %s will be applied to all requests", configuration.DefaultPolicy))

		return
	}

	for i, rule := range configuration.Rules {
		rulePosition := i + 1

		if len(rule.Domains) == 0 {
			validator.Push(fmt.Errorf("Rule #%d is invalid, a policy must have one or more domains", rulePosition))
		}

		if !IsPolicyValid(rule.Policy) {
			validator.Push(fmt.Errorf("Policy [%s] for rule #%d domain: %s is invalid, a policy must either be 'deny', 'two_factor', 'one_factor' or 'bypass'", rule.Policy, rulePosition, rule.Domains))
		}

		validateNetworks(rulePosition, rule, configuration, validator)

		validateResources(rulePosition, rule, validator)

		validateSubjects(rulePosition, rule, validator)

		validateMethods(rulePosition, rule, validator)

		if rule.Policy == policyBypass && len(rule.Subjects) != 0 {
			validator.Push(fmt.Errorf(errAccessControlInvalidPolicyWithSubjects, rulePosition, rule.Domains, rule.Subjects))
		}
	}
}

func validateNetworks(rulePosition int, rule schema.ACLRule, configuration schema.AccessControlConfiguration, validator *schema.StructValidator) {
	for _, network := range rule.Networks {
		if !IsNetworkValid(network) {
			if !IsNetworkGroupValid(configuration, network) {
				validator.Push(fmt.Errorf("Network %s for rule #%d domain: %s is not a valid network or network group", rule.Networks, rulePosition, rule.Domains))
			}
		}
	}
}

func validateResources(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	for _, resource := range rule.Resources {
		if err := IsResourceValid(resource); err != nil {
			validator.Push(fmt.Errorf("Resource %s for rule #%d domain: %s is invalid, %s", rule.Resources, rulePosition, rule.Domains, err))
		}
	}
}

func validateSubjects(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	for _, subjectRule := range rule.Subjects {
		for _, subject := range subjectRule {
			if !IsSubjectValid(subject) {
				validator.Push(fmt.Errorf("Subject %s for rule #%d domain: %s is invalid, must start with 'user:' or 'group:'", subjectRule, rulePosition, rule.Domains))
			}
		}
	}
}

func validateMethods(rulePosition int, rule schema.ACLRule, validator *schema.StructValidator) {
	for _, method := range rule.Methods {
		if !utils.IsStringInSliceFold(method, validHTTPRequestMethods) {
			validator.Push(fmt.Errorf("Method %s for rule #%d domain: %s is invalid, must be one of the following methods: %s", method, rulePosition, rule.Domains, strings.Join(validHTTPRequestMethods, ", ")))
		}
	}
}
