package compose

type Services struct {
	Version     string             `yaml:"version"`
	ServicesMap map[string]Service `yaml:"services"`
}

type Service struct {
	Image         string   `yaml:"image"`
	ContainerName string   `yaml:"container_name,omitempty"`
	Restart       string   `yaml:"restart,omitempty"`
	Environment   []string `yaml:"environment,omitempty"`
}
