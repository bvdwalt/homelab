package config

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// Config holds Docker connection configuration
type Config struct {
	DockerUsername         string
	DockerHostname         string
	DomainName             string
	SSDPath                string
	HDDPath                string
	ExternalPath           string
	BeszelKey              pulumi.StringOutput
	PostgresDbHost         string
	PostgresDbPassword     pulumi.StringOutput
	LinkwardenNextURL      string
	LinkwardenNextSecret   pulumi.StringOutput
	MinifluxDbName         string
	MinifluxDbUsername     string
	MinifluxdbUserPassword pulumi.StringOutput
	MinifluxAdminUsername  string
	MinifluxAdminPassword  pulumi.StringOutput
	MinifluxRunMigrations  string
	MinifluxCreateAdmin    string
	MinifluxDebug          bool
}

// GetConfig retrieves configuration from Pulumi config
func GetConfig(ctx *pulumi.Context) (*Config, error) {
	cfg := config.New(ctx, "homelab")

	username := cfg.Require("dockerUsername")
	hostname := cfg.Require("dockerHostname")
	domainName := cfg.Require("domain")
	ssdPath := cfg.Require("pathsSsd")
	hddPath := cfg.Require("pathsHdd")
	externalPath := cfg.Require("pathsExternal")
	beszelKey := cfg.RequireSecret("beszelKey")
	postgresDbHost := cfg.Require("linkwardenDbHost")
	postgresDbPassword := cfg.RequireSecret("linkwardenDbPassword")
	LinkwardenNextURL := cfg.Require("linkwardenNextAuthUrl")
	LinkwardenNextSecret := cfg.RequireSecret("linkwardenNextAuthSecret")
	minifluxDbName := cfg.Require("minifluxDbName")
	minifluxDbUsername := cfg.Require("minifluxDbUserName")
	minifluxdbUserPassword := cfg.RequireSecret("minifluxDbUserPassword")
	minifluxAdminUsername := cfg.Require("minifluxAdminUsername")
	minifluxAdminPassword := cfg.RequireSecret("minifluxAdminPassword")
	minifluxRunMigrations := cfg.Require("minifluxRunMigrations")
	minifluxCreateAdmin := cfg.Require("minifluxCreateAdmin")
	minifluxDebug := cfg.GetBool("minifluxDebug")

	return &Config{
		DockerUsername:         username,
		DockerHostname:         hostname,
		DomainName:             domainName,
		SSDPath:                ssdPath,
		HDDPath:                hddPath,
		ExternalPath:           externalPath,
		BeszelKey:              beszelKey,
		PostgresDbHost:         postgresDbHost,
		PostgresDbPassword:     postgresDbPassword,
		LinkwardenNextURL:      LinkwardenNextURL,
		LinkwardenNextSecret:   LinkwardenNextSecret,
		MinifluxDbName:         minifluxDbName,
		MinifluxDbUsername:     minifluxDbUsername,
		MinifluxdbUserPassword: minifluxdbUserPassword,
		MinifluxAdminUsername:  minifluxAdminUsername,
		MinifluxAdminPassword:  minifluxAdminPassword,
		MinifluxRunMigrations:  minifluxRunMigrations,
		MinifluxCreateAdmin:    minifluxCreateAdmin,
		MinifluxDebug:          minifluxDebug,
	}, nil
}

// SSHConnectionString returns the SSH connection string for Docker
func (c *Config) SSHConnectionString() string {
	return fmt.Sprintf("ssh://%s@%s", c.DockerUsername, c.DockerHostname)
}
