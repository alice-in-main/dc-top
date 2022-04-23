package logger

import (
	"fmt"
	"log"
	"os"
)

var log_file_name string

func Init() {
	// log_file_name = fmt.Sprintf("/tmp/dc-top-logs-%s.txt", utils.RandSeq(6))
	log_file_name = fmt.Sprintf("/tmp/dc-top-logs.txt")
	os.Remove(log_file_name)
	log_file, err := os.OpenFile(log_file_name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(log_file)
	log.Println("Hello world!")
}

func Cleanup() {
	// os.Remove(log_file_name)
}
