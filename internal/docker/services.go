package docker

import "path/filepath"

// HomelabServices contains pre-configured container definitions for common homelab services
type HomelabServices struct {
	DomainName   string
	Images       *ImageConfig
	SSDPath      string
	HDDPath      string
	ExternalPath string
}

// NewHomelabServices creates a new homelab services configuration
func NewHomelabServices(domainName string, images *ImageConfig, ssdPath string, hddPath string, externalPath string) *HomelabServices {
	return &HomelabServices{
		DomainName:   domainName,
		Images:       images,
		SSDPath:      ssdPath,
		HDDPath:      hddPath,
		ExternalPath: externalPath,
	}
}

// Whoami returns configuration for the Traefik whoami service
func (h *HomelabServices) Whoami() ContainerConfig {
	return ContainerConfig{
		Name:         "whoami",
		ImageName:    h.Images.Images["traefik-whoami"],
		InternalPort: 80,
		DomainName:   h.DomainName,
		ServiceName:  "whoami",
		Networks:     []string{"proxy"},
	}
}

// Beszel returns configuration for the Beszel monitoring dashboard
func (h *HomelabServices) Beszel() ContainerConfig {
	return ContainerConfig{
		Name:         "beszel",
		ImageName:    h.Images.Images["beszel"],
		InternalPort: 8090,
		DomainName:   h.DomainName,
		ServiceName:  "beszel",
		Networks:     []string{"proxy"},
		Volumes: []VolumeMount{
			{
				HostPath:      filepath.Join(h.SSDPath, "docker-volumes/beszel-data"),
				ContainerPath: "/beszel_data",
				ReadOnly:      false,
			},
			{
				HostPath:      filepath.Join(h.SSDPath, "docker-volumes/beszel-socket"),
				ContainerPath: "/beszel_socket",
				ReadOnly:      false,
			},
		},
		ExtraLabels: map[string]string{
			"traefik.http.routers.beszel.middlewares": "oidc-auth",
		},
	}
}

// BeszelAgent returns configuration for the Beszel monitoring agent
func (h *HomelabServices) BeszelAgent() ContainerConfig {
	return ContainerConfig{
		Name:        "beszel-agent",
		ImageName:   h.Images.Images["beszel-agent"],
		NetworkMode: "host",
		Volumes: []VolumeMount{
			{
				HostPath:      filepath.Join(h.SSDPath, "docker-volumes/beszel-socket"),
				ContainerPath: "/beszel_socket",
				ReadOnly:      false,
			},
			{
				HostPath:      "/var/run/docker.sock",
				ContainerPath: "/var/run/docker.sock",
				ReadOnly:      true,
			},
			{
				HostPath:      filepath.Join(h.SSDPath, ".beszel"),
				ContainerPath: "/extra-filesystems/Cheetah",
				ReadOnly:      true,
			},
			{
				HostPath:      filepath.Join(h.HDDPath, ".beszel"),
				ContainerPath: "/extra-filesystems/Hare",
				ReadOnly:      true,
			},
			{
				HostPath:      filepath.Join(h.ExternalPath, ".beszel"),
				ContainerPath: "/extra-filesystems/external",
				ReadOnly:      true,
			},
		},
		Environment: map[string]string{
			"LISTEN": "/beszel_socket/beszel.sock",
		},
	}
}
