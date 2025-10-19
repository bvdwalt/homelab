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
		config, err := config.GetConfig()
		if err != nil {
			return err
		}

		dockerProvider, err := dockerSDK.NewProvider(ctx, "RemoteDocker", &dockerSDK.ProviderArgs{
			Host: pulumi.String(config.SSHConnectionString()),
		})
		if err != nil {
			return err
		}
		ctx.Export("dockerProvider", dockerProvider.ID())

		whoamiImage, err := dockerSDK.NewRemoteImage(ctx, "whoami", &dockerSDK.RemoteImageArgs{
			Name: pulumi.String("traefik/whoami:latest"),
		}, pulumi.Provider(dockerProvider))
		if err != nil {
			return err
		}

		_, err = dockerSDK.NewContainer(ctx, "whoami2", &dockerSDK.ContainerArgs{
			Image: whoamiImage.Name,
			Ports: dockerSDK.ContainerPortArray{
				&dockerSDK.ContainerPortArgs{
					Internal: pulumi.Int(80),
					External: pulumi.Int(8181),
				},
			},
			Labels:           docker.TraefikLabels("whoami2", config.DomainName, 80),
			NetworksAdvanced: docker.ProxyNetworkConfig(),
		}, pulumi.Provider(dockerProvider))
		if err != nil {
			return err
		}
		return nil
	})
}
