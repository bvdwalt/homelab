package docker

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ImageConfig holds the Docker image versions loaded from YAML
type ImageConfig struct {
	Images map[string]string `yaml:"images"`
}

// LoadImageConfig loads Docker image versions from image-versions.yaml
func LoadImageConfig() (*ImageConfig, error) {
	data, err := os.ReadFile("image-versions.yaml")
	if err != nil {
		return nil, err
	}

	var config ImageConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
