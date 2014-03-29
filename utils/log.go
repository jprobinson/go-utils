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
	FileHandle() io.WriteCloser
}

func GetLogFileHandle(fileName string) io.WriteCloser {
	logFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Sprintf("Unable to get a hold of log file: %s", fileName))
	}
	return logFile
}

type DefaultLogSetup struct {
	logFile  io.WriteCloser
	fileName string
}

func NewDefaultLogSetup(logFileName string) *DefaultLogSetup {
	l := &DefaultLogSetup{fileName: logFileName}
	return l
}

func (l *DefaultLogSetup) FileHandle() io.WriteCloser {
	return l.logFile
}

// SetupLogging attempts to get a handle on the given log file then sets it to the log output.
// This function will also setup the log flags to include the go file name and line number.
func (l *DefaultLogSetup) SetupLogging() {
	l.logFile = GetLogFileHandle(l.fileName)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(l.logFile)
	log.Print("set up logger")
}

// ListenForLogSignal listens for a SIGHUP signal from the OS. When it recieves the signal,
// it will call the LogConfig to setup logging again.
func ListenForLogSignal(logSetup LogSetup) {
	logSignal := make(chan os.Signal, 1)
	signal.Notify(logSignal, syscall.SIGHUP)
	for {
		<-logSignal
		oldLog := logSetup.FileHandle()
		logSetup.SetupLogging()
		oldLog.Close()
	}
}
