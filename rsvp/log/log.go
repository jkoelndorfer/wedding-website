package log

import (
	"log"
	"os"
)

var logger *log.Logger

func Logger() *log.Logger {
	if logger == nil {
		logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
	}
	return logger
}
