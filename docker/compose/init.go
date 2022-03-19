package compose

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

var (
	is_dc_mode_on bool
	workdir_path  string
	dc_yaml_path  string
)

func Init(ctx context.Context, workdir string, dc_file_path string) error {
	_, err := os.Stat(dc_file_path)
	if err != nil {
		fmt.Printf("didn't find file %s\n", dc_file_path)
		panic(err)
	}
	is_dc_mode_on = true
	if dc_file_path[0] == '/' || dc_file_path[0] == '~' {
		dc_yaml_path = dc_file_path
	} else {
		curr_path, err := os.Getwd()
		if err != nil {
			fmt.Print("Failed to get current path on compose\n")
			panic(err)
		}
		dc_yaml_path, err = filepath.Abs(fmt.Sprintf("%s/%s", curr_path, dc_file_path))
		if err != nil {
			fmt.Print("Failed to calculate absolute path\n")
			panic(err)
		}
	}
	if !ValidateYaml(ctx) {
		return fmt.Errorf("failed to validate docker-compose yaml: '%s'. (run docker-compose config)", dc_file_path)
	}
	workdir_path = workdir
	return nil
}
