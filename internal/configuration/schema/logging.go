package schema

// LogConfiguration represents the logging configuration.
type LogConfiguration struct {
	Level      string `koanf:"level"`
	Format     string `koanf:"format"`
	FilePath   string `koanf:"file_path"`
	KeepStdout bool   `koanf:"keep_stdout"`
}

// DefaultLoggingConfiguration is the default logging configuration.
var DefaultLoggingConfiguration = LogConfiguration{
	Level:  "info",
	Format: "text",
}
