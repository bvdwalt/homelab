package config

import (
	"fmt"
	"os"
)

// Config holds Docker connection configuration
type Config struct {
	DockerUsername string
	DockerHostname string
	DomainName     string
}

// GetConfig retrieves configuration from environment variables
func GetConfig() (*Config, error) {
	username := os.Getenv("DOCKER_USERNAME")
	if username == "" {
		return nil, fmt.Errorf("DOCKER_USERNAME environment variable is not set")
	}

	hostname := os.Getenv("DOCKER_HOSTNAME")
	if hostname == "" {
		return nil, fmt.Errorf("DOCKER_HOSTNAME environment variable is not set")
	}

	domainName := os.Getenv("DOMAIN_NAME")
	if domainName == "" {
		return nil, fmt.Errorf("DOMAIN_NAME environment variable is not set")
	}

	return &Config{
		DockerUsername: username,
		DockerHostname: hostname,
		DomainName:     domainName,
	}, nil
}

// SSHConnectionString returns the SSH connection string for Docker
func (c *Config) SSHConnectionString() string {
	return fmt.Sprintf("ssh://%s@%s", c.DockerUsername, c.DockerHostname)
}
