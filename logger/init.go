package logger

import (
	"dc-top/utils"
	"fmt"
	"log"
	"os"
)

var (
	err           error
	log_file_name string
	log_file      *os.File
)

func Init() {
	log_file_name = fmt.Sprintf(`%s/dc-top-logs-%s.txt`, utils.TempFolderPath(), utils.RandSeq(6))
	os.Remove(log_file_name)
	log_file, err = os.OpenFile(log_file_name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(log_file)
	log.Println("Hello world!")
}

func Cleanup() {
	log_file.Close()
	if err = os.Remove(log_file_name); err != nil {
		fmt.Printf("Failed to remove log file %s. Error: %s", log_file_name, err.Error())
	}
}
