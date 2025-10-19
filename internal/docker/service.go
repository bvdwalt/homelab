package docker

import (
	"github.com/pulumi/pulumi-docker/sdk/v4/go/docker"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ContainerService manages Docker containers and images
type ContainerService struct {
	provider pulumi.ProviderResource
	ctx      *pulumi.Context
}

// NewContainerService creates a new container service
func NewContainerService(ctx *pulumi.Context, provider pulumi.ProviderResource) *ContainerService {
	return &ContainerService{
		provider: provider,
		ctx:      ctx,
	}
}

// ContainerConfig holds configuration for creating a container
type ContainerConfig struct {
	Name         string
	ImageName    string
	InternalPort int
	ExternalPort int
	DomainName   string
	ServiceName  string
	// Optional fields
	Environment   map[string]string
	Volumes       []VolumeMount
	Networks      []string
	RestartPolicy string
	ExtraLabels   map[string]string
}

// VolumeMount represents a volume mount configuration
type VolumeMount struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
}

// CreateContainer creates a container with its image using the provided configuration
func (cs *ContainerService) CreateContainer(config ContainerConfig) (*docker.Container, error) {
	image, err := cs.createImage(config.Name, config.ImageName)
	if err != nil {
		return nil, err
	}

	// Build container arguments
	containerArgs := &docker.ContainerArgs{
		Image: image.Name,
		Name:  pulumi.String(config.Name),
	}

	// Add ports if specified
	if config.InternalPort > 0 && config.ExternalPort > 0 {
		containerArgs.Ports = docker.ContainerPortArray{
			&docker.ContainerPortArgs{
				Internal: pulumi.Int(config.InternalPort),
				External: pulumi.Int(config.ExternalPort),
			},
		}
	}

	// Add Traefik labels if service name and domain are provided
	if config.ServiceName != "" && config.DomainName != "" {
		labels := TraefikLabels(config.ServiceName, config.DomainName, config.InternalPort)

		// Add extra labels if provided
		if len(config.ExtraLabels) > 0 {
			for key, value := range config.ExtraLabels {
				labels = append(labels, &docker.ContainerLabelArgs{
					Label: pulumi.String(key),
					Value: pulumi.String(value),
				})
			}
		}

		containerArgs.Labels = labels
	}

	// Add networks
	if len(config.Networks) > 0 {
		networks := make(docker.ContainerNetworksAdvancedArray, len(config.Networks))
		for i, network := range config.Networks {
			networks[i] = &docker.ContainerNetworksAdvancedArgs{
				Name: pulumi.String(network),
			}
		}
		containerArgs.NetworksAdvanced = networks
	} else {
		// Default to proxy network
		containerArgs.NetworksAdvanced = ProxyNetworkConfig()
	}

	// Add environment variables
	if len(config.Environment) > 0 {
		envVars := make(pulumi.StringArray, 0, len(config.Environment))
		for key, value := range config.Environment {
			envVars = append(envVars, pulumi.Sprintf("%s=%s", key, value))
		}
		containerArgs.Envs = envVars
	}

	// Add volumes
	if len(config.Volumes) > 0 {
		volumes := make(docker.ContainerVolumeArray, len(config.Volumes))
		for i, volume := range config.Volumes {
			volumes[i] = &docker.ContainerVolumeArgs{
				HostPath:      pulumi.String(volume.HostPath),
				ContainerPath: pulumi.String(volume.ContainerPath),
				ReadOnly:      pulumi.Bool(volume.ReadOnly),
			}
		}
		containerArgs.Volumes = volumes
	}

	// Set restart policy
	if config.RestartPolicy != "" {
		containerArgs.Restart = pulumi.String(config.RestartPolicy)
	} else {
		containerArgs.Restart = pulumi.String("unless-stopped")
	}

	// Create the container
	container, err := docker.NewContainer(cs.ctx, config.Name, containerArgs, pulumi.Provider(cs.provider))
	if err != nil {
		return nil, err
	}

	return container, nil
}

// createImage creates a Docker image
func (cs *ContainerService) createImage(resourceName, imageName string) (*docker.RemoteImage, error) {
	image, err := docker.NewRemoteImage(cs.ctx, resourceName+"-image", &docker.RemoteImageArgs{
		Name: pulumi.String(imageName),
	}, pulumi.Provider(cs.provider))
	if err != nil {
		return nil, err
	}

	return image, nil
}
