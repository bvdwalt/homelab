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

	return &Config{
		DockerUsername:         cfg.Require("dockerUsername"),
		DockerHostname:         cfg.Require("dockerHostname"),
		DomainName:             cfg.Require("domain"),
		SSDPath:                cfg.Require("pathsSsd"),
		HDDPath:                cfg.Require("pathsHdd"),
		ExternalPath:           cfg.Require("pathsExternal"),
		BeszelKey:              cfg.RequireSecret("beszelKey"),
		PostgresDbHost:         cfg.Require("linkwardenDbHost"),
		PostgresDbPassword:     cfg.RequireSecret("linkwardenDbPassword"),
		LinkwardenNextURL:      cfg.Require("linkwardenNextAuthUrl"),
		LinkwardenNextSecret:   cfg.RequireSecret("linkwardenNextAuthSecret"),
		MinifluxDbName:         cfg.Require("minifluxDbName"),
		MinifluxDbUsername:     cfg.Require("minifluxDbUserName"),
		MinifluxdbUserPassword: cfg.RequireSecret("minifluxDbUserPassword"),
		MinifluxAdminUsername:  cfg.Require("minifluxAdminUsername"),
		MinifluxAdminPassword:  cfg.RequireSecret("minifluxAdminPassword"),
		MinifluxRunMigrations:  cfg.Require("minifluxRunMigrations"),
		MinifluxCreateAdmin:    cfg.Require("minifluxCreateAdmin"),
		MinifluxDebug:          cfg.GetBool("minifluxDebug"),
	}, nil
}

// SSHConnectionString returns the SSH connection string for Docker
func (c *Config) SSHConnectionString() string {
	return fmt.Sprintf("ssh://%s@%s", c.DockerUsername, c.DockerHostname)
}
