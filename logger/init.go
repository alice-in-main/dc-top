package logger

import (
	"log"
	"os"
)

func Init() {
	log_file_name := "/tmp/dc-top-logs.txt"
	os.Remove(log_file_name)
	log_file, err := os.OpenFile(log_file_name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(log_file)
	log.Println("Hello world!")
}
