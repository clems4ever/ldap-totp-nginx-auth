package schema

// AccessControlConfiguration represents the configuration related to ACLs.
type AccessControlConfiguration struct {
	DefaultPolicy string       `koanf:"default_policy"`
	Networks      []ACLNetwork `koanf:"networks"`
	Rules         []ACLRule    `koanf:"rules"`
}

// ACLNetwork represents one ACL network group entry; "weak" coerces a single value into slice.
type ACLNetwork struct {
	Name     string   `koanf:"name"`
	Networks []string `koanf:"networks"`
}

// ACLRule represents one ACL rule entry; "weak" coerces a single value into slice.
type ACLRule struct {
	Domains   []string   `koanf:"domain"`
	Policy    string     `koanf:"policy"`
	Subjects  [][]string `koanf:"subject"`
	Networks  []string   `koanf:"networks"`
	Resources []string   `koanf:"resources"`
	Methods   []string   `koanf:"methods"`
}

// DefaultACLNetwork represents the default configuration related to access control network group configuration.
var DefaultACLNetwork = []ACLNetwork{
	{
		Name:     "localhost",
		Networks: []string{"127.0.0.1"},
	},
	{
		Name:     "internal",
		Networks: []string{"10.0.0.0/8"},
	},
}

// DefaultACLRule represents the default configuration related to access control rule configuration.
var DefaultACLRule = []ACLRule{
	{
		Domains: []string{"public.example.com"},
		Policy:  "bypass",
	},
	{
		Domains: []string{"singlefactor.example.com"},
		Policy:  "one_factor",
	},
	{
		Domains: []string{"secure.example.com"},
		Policy:  "two_factor",
	},
}
