package logger

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	os.Remove("app.log")
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	Logger = log.New(file, "PREFIX: ", log.Ldate|log.Ltime|log.Lshortfile)
}
