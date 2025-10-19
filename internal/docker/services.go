package docker

// HomelabServices contains pre-configured container definitions for common homelab services
type HomelabServices struct {
	DomainName string
}

// NewHomelabServices creates a new homelab services configuration
func NewHomelabServices(domainName string) *HomelabServices {
	return &HomelabServices{
		DomainName: domainName,
	}
}

// Whoami returns configuration for the Traefik whoami service
func (h *HomelabServices) Whoami() ContainerConfig {
	return ContainerConfig{
		Name:         "whoami",
		ImageName:    "traefik/whoami:latest",
		InternalPort: 80,
		DomainName:   h.DomainName,
		ServiceName:  "whoami",
		Networks:     []string{"proxy"},
	}
}
