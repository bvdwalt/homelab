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
	SSDPath        string
	HDDPath        string
	ExternalPath   string
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

	ssdPath := os.Getenv("DOCKER_SSD_ROOT_PATH")
	if ssdPath == "" {
		return nil, fmt.Errorf("DOCKER_SSD_ROOT_PATH environment variable is not set")
	}

	hddPath := os.Getenv("DOCKER_HDD_ROOT_PATH")
	if hddPath == "" {
		return nil, fmt.Errorf("DOCKER_HDD_ROOT_PATH environment variable is not set")
	}

	externalPath := os.Getenv("DOCKER_EXTERNAL_ROOT_PATH")
	if externalPath == "" {
		return nil, fmt.Errorf("DOCKER_EXTERNAL_ROOT_PATH environment variable is not set")
	}

	return &Config{
		DockerUsername: username,
		DockerHostname: hostname,
		DomainName:     domainName,
		SSDPath:        ssdPath,
		HDDPath:        hddPath,
		ExternalPath:   externalPath,
	}, nil
}

// SSHConnectionString returns the SSH connection string for Docker
func (c *Config) SSHConnectionString() string {
	return fmt.Sprintf("ssh://%s@%s", c.DockerUsername, c.DockerHostname)
}
