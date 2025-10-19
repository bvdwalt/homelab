package main

import (
	"homelab/internal/config"
	"homelab/internal/docker"

	dockerSDK "github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg, err := config.GetConfig()
		if err != nil {
			return err
		}

		dockerProvider, err := dockerSDK.NewProvider(ctx, "RemoteDocker", &dockerSDK.ProviderArgs{
			Host: pulumi.String(cfg.SSHConnectionString()),
		})
		if err != nil {
			return err
		}
		ctx.Export("dockerProvider", dockerProvider.ID())

		containerService := docker.NewContainerService(ctx, dockerProvider)

		homelabServices := docker.NewHomelabServices(cfg.DomainName)

		services := []docker.ContainerConfig{
			homelabServices.Whoami(),
		}

		for _, serviceConfig := range services {
			_, err := containerService.CreateContainer(serviceConfig)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
