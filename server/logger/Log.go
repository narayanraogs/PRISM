package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var Log *slog.Logger
var logFile *os.File

func InitializeLog(path string) {
	now := time.Now().Format("2006-01-02-15-04-05")
	fileName := filepath.Join(path, "prism-"+now+".log")
	var err error
	logFile, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Failed to open log file", err.Error())
		return
	}
	Log = slog.New(slog.NewTextHandler(logFile, nil))
	fmt.Println("New Log Started")
}

func CloseLog() {
	err := logFile.Close()
	if err != nil {
		fmt.Println("Unable to close Log file")
		return
	}
	fmt.Println("Log File Closed")
}
