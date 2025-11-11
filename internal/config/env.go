package config

import (
	"fmt"
	"os"
)

// Config holds Docker connection configuration
type Config struct {
	DockerUsername       string
	DockerHostname       string
	DomainName           string
	SSDPath              string
	HDDPath              string
	ExternalPath         string
	BeszelKey            string
	LinkwardenDbHost     string
	LinkwardenDbName     string
	LinkwardenDbPassword string
	LinkwardenNextURL    string
	LinkwardenNextSecret string
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

	beszelKey := os.Getenv("BESZEL_KEY")
	if beszelKey == "" {
		return nil, fmt.Errorf("BESZEL_KEY environment variable is not set")
	}

	LinkwardenDbHost := os.Getenv("LINKWARDEN_DBHOST")
	if LinkwardenDbHost == "" {
		return nil, fmt.Errorf("LINKWARDEN_DBHOST environment variable is not set")
	}

	LinkwardenDbName := os.Getenv("LINKWARDEN_DBNAME")
	if LinkwardenDbName == "" {
		return nil, fmt.Errorf("LINKWARDEN_DBNAME environment variable is not set")
	}

	LinkwardenDbPassword := os.Getenv("LINKWARDEN_DBPASSWORD")
	if LinkwardenDbPassword == "" {
		return nil, fmt.Errorf("LINKWARDEN_DBPASSWORD environment variable is not set")
	}

	LinkwardenNextURL := os.Getenv("LINKWARDEN_NEXTAUTH_URL")
	if LinkwardenNextURL == "" {
		return nil, fmt.Errorf("LINKWARDEN_NEXTAUTH_URL environment variable is not set")
	}

	LinkwardenNextSecret := os.Getenv("LINKWARDEN_NEXTAUTH_SECRET")
	if LinkwardenNextSecret == "" {
		return nil, fmt.Errorf("LINKWARDEN_NEXTAUTH_SECRET environment variable is not set")
	}

	return &Config{
		DockerUsername:       username,
		DockerHostname:       hostname,
		DomainName:           domainName,
		SSDPath:              ssdPath,
		HDDPath:              hddPath,
		ExternalPath:         externalPath,
		BeszelKey:            beszelKey,
		LinkwardenDbHost:     LinkwardenDbHost,
		LinkwardenDbName:     LinkwardenDbName,
		LinkwardenDbPassword: LinkwardenDbPassword,
		LinkwardenNextURL:    LinkwardenNextURL,
		LinkwardenNextSecret: LinkwardenNextSecret,
	}, nil
}

// SSHConnectionString returns the SSH connection string for Docker
func (c *Config) SSHConnectionString() string {
	return fmt.Sprintf("ssh://%s@%s", c.DockerUsername, c.DockerHostname)
}
