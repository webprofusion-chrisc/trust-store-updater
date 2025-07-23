package certstore

import (
	"fmt"
	"os"
)

type LogLevel int

const (
	LogError LogLevel = iota
	LogWarn
	LogInfo
)

var currentLevel = LogInfo

func SetLogLevel(level LogLevel) {
	currentLevel = level
}

func LogInfof(format string, args ...interface{}) {
	if currentLevel >= LogInfo {
		fmt.Fprintf(os.Stdout, "INFO: "+format+"\n", args...)
	}
}

func LogWarnf(format string, args ...interface{}) {
	if currentLevel >= LogWarn {
		fmt.Fprintf(os.Stdout, "WARN: "+format+"\n", args...)
	}
}

func LogErrorf(format string, args ...interface{}) {
	if currentLevel >= LogError {
		fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	}
}
