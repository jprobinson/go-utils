package log

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
)

type LogSetup interface {
    SetupLogging()
}

type DefaultLogSetup struct {
    LogFile string
}

func GetLogFileHandle(log_file_name string) *os.File {
    log_file, err := os.OpenFile(log_file_name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
    if err != nil {
        panic(fmt.Sprintf("Unable to get a hold of log file: %s", log_file_name))
    }
    return log_file
}

// SetupLogging attempts to get a handle on the given log file then sets it to the log output.
// This function will also setup the log flags to include the go file name and line number.
func (log_setup DefaultLogSetup) SetupLogging() {
    log_file := GetLogFileHandle(log_setup.LogFile)
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    log.SetOutput(log_file)
    log.Print("setup default logger")
}

// ListenForLogSignal listens for a SIGHUP signal from the OS. When it recieves the signal,
// it will call the LogConfig to setup logging again.
func ListenForLogSignal(log_setup LogSetup) {
    log_signal := make(chan os.Signal, 1)
    signal.Notify(log_signal, syscall.SIGHUP)
    for {
        <-log_signal
        log_setup.SetupLogging()
    }
}

