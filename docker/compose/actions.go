package compose

import (
	"context"
	"dc-top/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

// TODO: add --compatability
func Up(ctx context.Context) ([]byte, error) {
	return exec.CommandContext(ctx, "docker-compose", "--compatibility", "-f", DcYamlPath(), "up").CombinedOutput()
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
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", DcYamlPath(), "config", "-q")
	err := cmd.Run()
	return err == nil
}

func Config(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", DcYamlPath(), "config", "-q")
	return cmd.CombinedOutput()
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

func GetDcProcesses(ctx context.Context) ([]Process, error) {
	raw_processes, err := exec.Command("docker", "compose", "-f", DcYamlPath(), "ps", "--format", "json").Output()
	if err != nil {
		log.Println("failed to get docker compose data")
		return nil, err
	}

	var parsed_processes []Process
	json.Unmarshal(raw_processes, &parsed_processes)

	return parsed_processes, nil
}

func backupFileName() string {
	return fmt.Sprintf("%s/%s.backup", workdir_path, ".docker-compose.yaml")
}
