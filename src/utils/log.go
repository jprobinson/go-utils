package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type LogSetup interface {
	SetupLogging()
}

// ListenForLogSignal listens for a SIGHUP signal from the OS. When it recieves the signal,
// it will call the l to setup logging again.
func ListenForLogSignal(l LogSetup) {
	logSignal := make(chan os.Signal, 1)
	signal.Notify(logSignal, syscall.SIGHUP)
	for {
		<-logSignal
		l.SetupLogging()
	}
}

type DefaultLogSetup struct {
	LogFile string
}

// SetupLogging attempts to get a handle on the given log file then sets it to the log output.
// This function will also setup the log flags to include the go file name and line number.
func (l DefaultLogSetup) SetupLogging() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(GetLogFileHandle(l.LogFile))
	log.Print("setup default logger")
}

func GetLogFileHandle(logFileName string) io.Writer {
	logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		panic(fmt.Sprintf("Unable to get a hold of log file: %s", logFile))
	}
	return logFile
}
