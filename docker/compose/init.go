package compose

import (
	"context"
	"dc-top/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	is_dc_mode_on bool
	workdir_path  string
	dc_yaml_path  string
)

var workdir string

func Init(ctx context.Context, dc_file_path string) error {
	workdir = fmt.Sprintf("%s/dc-top-files-%s", utils.TempFolderPath(), utils.RandSeq(6))

	err := os.Mkdir(workdir, 0755)
	if err != nil {
		return err
	}
	defer os.RemoveAll(workdir)

	_, err = os.Stat(dc_file_path)
	if err != nil {
		log.Printf("didn't find file %s\n", dc_file_path)
		return err
	}

	is_dc_mode_on = true
	if dc_file_path[0] == '/' || dc_file_path[0] == '~' {
		dc_yaml_path = dc_file_path
	} else {
		curr_path, err := os.Getwd()
		if err != nil {
			log.Print("Failed to get current path on compose\n")
			return err
		}
		dc_yaml_path, err = filepath.Abs(fmt.Sprintf("%s/%s", curr_path, dc_file_path))
		if err != nil {
			log.Print("Failed to calculate absolute path\n")
			return err
		}
	}

	if !ValidateYaml(ctx) {
		return fmt.Errorf("failed to validate docker-compose yaml: '%s'. (run docker-compose config)", dc_file_path)
	}

	err = UpdateContainerFilters(ctx)
	if err != nil {
		return fmt.Errorf("failed to get initialize filters: '%s", err)
	}

	workdir_path = workdir
	return nil
}

func Cleanup() {
	os.RemoveAll(workdir)
}
