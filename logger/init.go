package logger

import (
	"fmt"
	"log"
	"os"
)

func Init(workdir string) {
	log_file_name := fmt.Sprintf("%s/dc-top-logs.txt", workdir)
	os.Remove(log_file_name)
	log_file, err := os.OpenFile(log_file_name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(log_file)
	log.Println("Hello world!")
}
