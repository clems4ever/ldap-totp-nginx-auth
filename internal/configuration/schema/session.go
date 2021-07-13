package schema

// RedisNode Represents a Node.
type RedisNode struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

// RedisHighAvailabilityConfiguration holds configuration variables for Redis Cluster/Sentinel.
type RedisHighAvailabilityConfiguration struct {
	SentinelName     string      `koanf:"sentinel_name"`
	SentinelPassword string      `koanf:"sentinel_password"`
	Nodes            []RedisNode `koanf:"nodes"`
	RouteByLatency   bool        `koanf:"route_by_latency"`
	RouteRandomly    bool        `koanf:"route_randomly"`
}

// RedisSessionConfiguration represents the configuration related to redis session store.
type RedisSessionConfiguration struct {
	Host                     string                              `koanf:"host"`
	Port                     int                                 `koanf:"port"`
	Username                 string                              `koanf:"username"`
	Password                 string                              `koanf:"password"`
	DatabaseIndex            int                                 `koanf:"database_index"`
	MaximumActiveConnections int                                 `koanf:"maximum_active_connections"`
	MinimumIdleConnections   int                                 `koanf:"minimum_idle_connections"`
	TLS                      *TLSConfig                          `koanf:"tls"`
	HighAvailability         *RedisHighAvailabilityConfiguration `koanf:"high_availability"`
}

// SessionConfiguration represents the configuration related to user sessions.
type SessionConfiguration struct {
	Name               string                     `koanf:"name"`
	Domain             string                     `koanf:"domain"`
	SameSite           string                     `koanf:"same_site"`
	Secret             string                     `koanf:"secret"`
	Expiration         string                     `koanf:"expiration"`
	Inactivity         string                     `koanf:"inactivity"`
	RememberMeDuration string                     `koanf:"remember_me_duration"`
	Redis              *RedisSessionConfiguration `koanf:"redis"`
}

// DefaultSessionConfiguration is the default session configuration.
var DefaultSessionConfiguration = SessionConfiguration{
	Name:               "authelia_session",
	Expiration:         "1h",
	Inactivity:         "5m",
	RememberMeDuration: "1M",
	SameSite:           "lax",
}
