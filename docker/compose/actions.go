package compose

import (
	"context"
	"dc-top/utils"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/api/types/filters"
	"gopkg.in/yaml.v2"
)

func Up(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker-compose", "-f", DcYamlPath(), "up")
	return cmd.Run()
}

func Down(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker-compose", "-f", DcYamlPath(), "down")
	return cmd.Run()
}

func Restart(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker-compose", "-f", DcYamlPath(), "restart")
	return cmd.Run()
}

func ValidateYaml(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "docker-compose", "-f", DcYamlPath(), "config", "-q")
	err := cmd.Run()
	return err == nil
}

func GenerateDcData(ctx context.Context) (Services, error) {
	if ValidateYaml(ctx) {
		var services Services
		contents, err := os.ReadFile(DcYamlPath())
		if err != nil {
			return services, err
		}
		err = yaml.Unmarshal(contents, &services)
		return services, err
	}
	return Services{}, errors.New("dc yaml is invalid")
}

func DcModeEnabled() bool {
	return is_dc_mode_on
}

func DcYamlPath() string {
	return dc_yaml_path
}

func IsYamlChanged() bool {
	file1, err := os.Stat(DcYamlPath())
	if err != nil {
		return true
	}
	file2, err := os.Stat(backupFileName())
	if err != nil {
		return true
	}
	return os.SameFile(file1, file2)
}

func CreateBackupYaml() error {
	return utils.CopyFile(DcYamlPath(), backupFileName())
}

func RestoreFromBackup() error {
	return utils.CopyFile(backupFileName(), DcYamlPath())
}

func CreateDcFilters(ctx context.Context) (*filters.Args, error) {
	dc_services, err := GenerateDcData(ctx)
	if err != nil {
		return nil, err
	}
	dc_filters := filters.NewArgs()
	for service_key, service := range dc_services.ServicesMap {
		if service.ContainerName == "" {
			dc_filters.Add("name", fmt.Sprintf("%s_%s", filepath.Base(filepath.Dir(DcYamlPath())), service_key))
		} else {
			dc_filters.Add("name", service.ContainerName)
		}
	}
	return &dc_filters, nil
}

func backupFileName() string {
	return fmt.Sprintf("%s/%s.backup", workdir_path, ".docker-compose.yaml")
}
