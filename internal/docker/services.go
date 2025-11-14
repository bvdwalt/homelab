package docker

import (
	"fmt"
	"path/filepath"
)

// HomelabServices contains pre-configured container definitions for common homelab services
type HomelabServices struct {
	DomainName   string
	Images       *ImageConfig
	SSDPath      string
	HDDPath      string
	ExternalPath string
}

// MinifluxSettings holds the optional environment values needed by the Miniflux container.
type MinifluxSettings struct {
	DatabaseHost         string
	DatabaseUserName     string
	DatabaseUserPassword string
	DatabaseName         string
	AdminUsername        string
	AdminPassword        string
	RunMigrations        string
	CreateAdmin          string
	Debug                bool
}

// NewHomelabServices creates a new homelab services configuration
func NewHomelabServices(domainName, ssdPath, hddPath, externalPath string, images *ImageConfig) *HomelabServices {
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

func (h *HomelabServices) Linkwarden(dbHost, dbPassword, nextUrl, nextSecret string) ContainerConfig {
	return ContainerConfig{
		Name:         "linkwarden",
		ServiceName:  "linkwarden",
		ImageName:    h.Images.Images["linkwarden"],
		InternalPort: 3000,
		DomainName:   h.DomainName,
		Networks:     []string{"proxy"},
		Volumes: []VolumeMount{
			{
				HostPath:      filepath.Join(h.SSDPath, "docker-volumes/linkwarden"),
				ContainerPath: "/data/data",
				ReadOnly:      false,
			},
		},
		Environment: map[string]string{
			"DATABASE_URL":    fmt.Sprintf("postgresql://postgres:%s@%s:5432/%s", dbPassword, dbHost, "linkwarden"),
			"NEXTAUTH_URL":    nextUrl,
			"NEXTAUTH_SECRET": nextSecret,
		},
		ExtraLabels: map[string]string{
			"traefik.http.routers.beszel.middlewares": "oidc-auth",
		},
	}
}

// Miniflux returns configuration for the Miniflux RSS reader
func (h *HomelabServices) Miniflux(settings MinifluxSettings) ContainerConfig {
	return ContainerConfig{
		Name:         "miniflux",
		ServiceName:  "miniflux",
		ImageName:    h.Images.Images["miniflux"],
		InternalPort: 8080,
		DomainName:   h.DomainName,
		Networks:     []string{"proxy"},
		Volumes: []VolumeMount{
			{
				HostPath:      filepath.Join(h.SSDPath, "docker-volumes/miniflux"),
				ContainerPath: "/var/lib/miniflux",
				ReadOnly:      false,
			},
		},
		Environment: map[string]string{
			"DATABASE_URL":      fmt.Sprintf("postgresql://%s:%s@%s:5432/%s?sslmode=disable", settings.DatabaseUserName, settings.DatabaseUserPassword, settings.DatabaseHost, settings.DatabaseName),
			"MINIFLUX_BASE_URL": fmt.Sprintf("miniflux.%v", h.DomainName),
			"RUN_MIGRATIONS":    settings.RunMigrations,
			"CREATE_ADMIN":      settings.CreateAdmin,
			"ADMIN_USERNAME":    settings.AdminUsername,
			"ADMIN_PASSWORD":    settings.AdminPassword,
			"DEBUG":             fmt.Sprintf("%t", settings.Debug),
		},
		ExtraLabels: map[string]string{
			"traefik.http.routers.miniflux.middlewares": "oidc-auth",
		},
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
func (h *HomelabServices) BeszelAgent(beszelKey string) ContainerConfig {
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
			{
				HostPath:      filepath.Join(h.SSDPath, "docker-volumes/beszel-agent"),
				ContainerPath: "/var/lib/beszel-agent",
				ReadOnly:      false,
			},
		},
		Environment: map[string]string{
			"LISTEN": "/beszel_socket/beszel.sock",
			"KEY":    beszelKey,
		},
	}
}
