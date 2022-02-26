package compose

import (
	"dc-top/utils"
	"fmt"
	"os"
	"path/filepath"
)

var (
	is_dc_mode_on           bool
	work_yaml_file_path     string
	original_yaml_file_path string
)

func Init(workdir string, dc_file_path string) {
	_, err := os.Stat(dc_file_path)
	if err != nil {
		fmt.Printf("didn't find file %s\n", dc_file_path)
		panic(err)
	}
	work_yaml_file_path = fmt.Sprintf("%s/docker-compose.yaml", workdir)
	err = utils.CopyFile(dc_file_path, work_yaml_file_path)
	if err != nil {
		fmt.Printf("failed to copy %s\n", dc_file_path)
		panic(err)
	}
	is_dc_mode_on = true
	if dc_file_path[0] == '/' || dc_file_path[0] == '~' {
		original_yaml_file_path = dc_file_path
	} else {
		curr_path, err := os.Getwd()
		if err != nil {
			fmt.Print("Failed to get current path on compose\n")
			panic(err)
		}
		original_yaml_file_path, err = filepath.Abs(fmt.Sprintf("%s/%s", curr_path, dc_file_path))
		if err != nil {
			fmt.Print("Failed to calculate absolute path\n")
			panic(err)
		}
	}
}

func DcModeEnabled() bool {
	return is_dc_mode_on
}

func DcYamlPath() string {
	return work_yaml_file_path
}

func OriginalDcYamlPath() string {
	return original_yaml_file_path
}
