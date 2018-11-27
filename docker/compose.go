package docker

import (
	"github.com/byuoitav/common/nerr"
	yaml "gopkg.in/yaml.v2"
)

// NewCompose returns a new docker compose struct
func NewCompose() Compose {
	return Compose{
		Version:  "3",
		Services: make(map[string]Service),
	}
}

// NewServiceWithDefaultOptions .
func NewServiceWithDefaultOptions(image string, ports, environment []string) Service {
	return Service{
		Image:       image,
		Ports:       ports,
		Environment: environment,
		NetworkMode: "host",
		Restart:     "always",
		TTY:         true,
		Logging: ServiceLogging{
			Options: LoggingOptions{
				MaxSize:       "100m",
				Mode:          "non-blocking",
				MaxBufferSize: "4m",
			},
		},
	}
}

// GetFileBytes .
func GetFileBytes(compose Compose) ([]byte, *nerr.E) {
	bytes, err := yaml.Marshal(compose)
	if err != nil {
		return nil, nerr.Translate(err).Addf("failed to marshal yaml")
	}

	return bytes, nil
}
