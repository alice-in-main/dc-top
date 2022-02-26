package compose

type Services struct {
	Version  string             `yaml:"version"`
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	ContainerName string   `yaml:"container_name"`
	Image         string   `yaml:"image"`
	Restart       string   `yaml:"restart,omitempty"`
	Environment   []string `yaml:"environment,omitempty"`
}
