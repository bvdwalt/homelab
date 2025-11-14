package config

import (
	"fmt"
	"os"
)

// Config holds Docker connection configuration
type Config struct {
	DockerUsername        string
	DockerHostname        string
	DomainName            string
	SSDPath               string
	HDDPath               string
	ExternalPath          string
	BeszelKey             string
	PostgresDbHost        string
	PostgresDbPassword    string
	LinkwardenNextURL     string
	LinkwardenNextSecret  string
	MinifluxDbName        string
	MinifluxDbUsername    string
	MinifluxAdminUsername string
	MinifluxAdminPassword string
	MinifluxRunMigrations string
	MinifluxCreateAdmin   string
	MinifluxDebug         bool
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

	postgresDbHost := os.Getenv("LINKWARDEN_DBHOST")
	if postgresDbHost == "" {
		return nil, fmt.Errorf("LINKWARDEN_DBHOST environment variable is not set")
	}

	postgresDbPassword := os.Getenv("LINKWARDEN_DBPASSWORD")
	if postgresDbPassword == "" {
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

	minifluxDbName := os.Getenv("MINIFLUX_DBNAME")
	if minifluxDbName == "" {
		return nil, fmt.Errorf("MINIFLUX_DBNAME environment variable is not set")
	}

	minifluxDbUsername := os.Getenv("MINIFLUX_DBUSERNAME")
	if minifluxDbUsername == "" {
		return nil, fmt.Errorf("MINIFLUX_DBUSERNAME environment variable is not set")
	}

	minifluxAdminUsername := os.Getenv("MINIFLUX_ADMIN_USERNAME")
	if minifluxAdminUsername == "" {
		return nil, fmt.Errorf("MINIFLUX_ADMIN_USERNAME environment variable is not set")
	}

	minifluxAdminPassword := os.Getenv("MINIFLUX_ADMIN_PASSWORD")
	if minifluxAdminPassword == "" {
		return nil, fmt.Errorf("MINIFLUX_ADMIN_PASSWORD environment variable is not set")
	}

	minifluxRunMigrations := os.Getenv("MINIFLUX_RUN_MIGRATIONS")
	if minifluxRunMigrations == "" {
		return nil, fmt.Errorf("MINIFLUX_RUN_MIGRATIONS environment variable is not set")
	}

	minifluxCreateAdmin := os.Getenv("MINIFLUX_CREATE_ADMIN")
	if minifluxCreateAdmin == "" {
		return nil, fmt.Errorf("MINIFLUX_CREATE_ADMIN environment variable is not set")
	}

	minifluxDebug := os.Getenv("MINIFLUX_DEBUG")
	if minifluxDebug == "" {
		return nil, fmt.Errorf("MINIFLUX_DEBUG environment variable is not set")
	}

	return &Config{
		DockerUsername:        username,
		DockerHostname:        hostname,
		DomainName:            domainName,
		SSDPath:               ssdPath,
		HDDPath:               hddPath,
		ExternalPath:          externalPath,
		BeszelKey:             beszelKey,
		PostgresDbHost:        postgresDbHost,
		PostgresDbPassword:    postgresDbPassword,
		LinkwardenNextURL:     LinkwardenNextURL,
		LinkwardenNextSecret:  LinkwardenNextSecret,
		MinifluxDbName:        minifluxDbName,
		MinifluxDbUsername:    minifluxDbUsername,
		MinifluxAdminUsername: minifluxAdminUsername,
		MinifluxAdminPassword: minifluxAdminPassword,
		MinifluxRunMigrations: minifluxRunMigrations,
		MinifluxCreateAdmin:   minifluxCreateAdmin,
		MinifluxDebug:         minifluxDebug == "1",
	}, nil
}

// SSHConnectionString returns the SSH connection string for Docker
func (c *Config) SSHConnectionString() string {
	return fmt.Sprintf("ssh://%s@%s", c.DockerUsername, c.DockerHostname)
}
