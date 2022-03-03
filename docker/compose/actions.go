package compose

import (
	"context"
	"dc-top/utils"
	"fmt"
	"os"
	"os/exec"

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

func GenerateDcData() (Services, error) {
	var services Services
	contents, err := os.ReadFile(DcYamlPath())
	if err != nil {
		return services, err
	}
	err = yaml.Unmarshal(contents, &services)
	return services, err
}

func DcModeEnabled() bool {
	return is_dc_mode_on
}

func DcYamlPath() string {
	return dc_yaml_path
}

func CreateBackupYaml() error {
	return utils.CopyFile(DcYamlPath(), backupFileName())
}

func RestoreFromBackup() error {
	return utils.CopyFile(backupFileName(), DcYamlPath())
}

func backupFileName() string {
	return fmt.Sprintf("%s/%s.backup", workdir_path, ".docker-compose.yaml")
}
