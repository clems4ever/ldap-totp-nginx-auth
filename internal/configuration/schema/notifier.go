package schema

// FileSystemNotifierConfiguration represents the configuration of the notifier writing emails in a file.
type FileSystemNotifierConfiguration struct {
	Filename string `koanf:"filename"`
}

// SMTPNotifierConfiguration represents the configuration of the SMTP server to send emails with.
type SMTPNotifierConfiguration struct {
	Host                string     `koanf:"host"`
	Port                int        `koanf:"port"`
	Username            string     `koanf:"username"`
	Password            string     `koanf:"password"`
	Identifier          string     `koanf:"identifier"`
	Sender              string     `koanf:"sender"`
	Subject             string     `koanf:"subject"`
	StartupCheckAddress string     `koanf:"startup_check_address"`
	DisableRequireTLS   bool       `koanf:"disable_require_tls"`
	DisableHTMLEmails   bool       `koanf:"disable_html_emails"`
	TLS                 *TLSConfig `koanf:"tls"`
}

// NotifierConfiguration represents the configuration of the notifier to use when sending notifications to users.
type NotifierConfiguration struct {
	DisableStartupCheck bool                             `koanf:"disable_startup_check"`
	FileSystem          *FileSystemNotifierConfiguration `koanf:"filesystem"`
	SMTP                *SMTPNotifierConfiguration       `koanf:"smtp"`
}

// DefaultSMTPNotifierConfiguration represents default configuration parameters for the SMTP notifier.
var DefaultSMTPNotifierConfiguration = SMTPNotifierConfiguration{
	Subject:    "[Authelia] {title}",
	Identifier: "localhost",
	TLS: &TLSConfig{
		MinimumVersion: "TLS1.2",
	},
}
