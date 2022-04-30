//go:build windows
// +build windows

package utils

import (
	"fmt"
	"log"
	"os/user"
)

func TempFolderPath() string {
	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	return fmt.Sprintf(`%s\AppData\Local\Temp`, user.HomeDir)
}
