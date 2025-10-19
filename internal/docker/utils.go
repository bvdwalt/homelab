package docker

import (
	"fmt"

	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ProxyNetworkConfig returns the configuration for connecting containers to the proxy network
func ProxyNetworkConfig() docker.ContainerNetworksAdvancedArray {
	return docker.ContainerNetworksAdvancedArray{
		&docker.ContainerNetworksAdvancedArgs{
			Name: pulumi.String("proxy"),
		},
	}
}

// TraefikLabels generates Traefik labels for container routing and load balancing
func TraefikLabels(serviceName, domainName string, port int) docker.ContainerLabelArray {
	return docker.ContainerLabelArray{
		&docker.ContainerLabelArgs{
			Label: pulumi.String("traefik.enable"),
			Value: pulumi.String("true"),
		},
		&docker.ContainerLabelArgs{
			Label: pulumi.String(fmt.Sprintf("traefik.http.routers.%s.rule", serviceName)),
			Value: pulumi.String(fmt.Sprintf("Host(`%s.%s`)", serviceName, domainName)),
		},
		&docker.ContainerLabelArgs{
			Label: pulumi.String(fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", serviceName)),
			Value: pulumi.String(fmt.Sprintf("%d", port)),
		},
	}
}
