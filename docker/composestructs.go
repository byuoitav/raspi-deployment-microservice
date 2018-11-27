package docker

// Compose represents a docker-compose file
type Compose struct {
	Version  string             `yaml:"version"`
	Services map[string]Service `yaml:"services"`
}

// A Service represents a docker-compose service
type Service struct {
	Image       string         `yaml:"image"`
	Ports       []string       `yaml:"ports"`
	Environment []string       `yaml:"environment"`
	NetworkMode string         `yaml:"network_mode"`
	Restart     string         `yaml:"restart"`
	TTY         bool           `yaml:"tty"`
	Logging     ServiceLogging `yaml:"logging"`
}

// ServiceLogging .
type ServiceLogging struct {
	Options LoggingOptions `yaml:"options"`
}

// LoggingOptions .
type LoggingOptions struct {
	MaxSize       string `yaml:"max-size"`
	Mode          string `yaml:"mode"`
	MaxBufferSize string `yaml:"max-buffer-size"`
}
