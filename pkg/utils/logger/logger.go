package logger

import (
	"log"
)

type Level uint8

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarningLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

var Access Level = TraceLevel

func Load(level string) {
	switch level {
	case "panic":
		Access = PanicLevel
	case "fatal":
		Access = FatalLevel
	case "error":
		Access = ErrorLevel
	case "warning":
		Access = WarningLevel
	case "info":
		Access = InfoLevel
	case "debug":
		Access = DebugLevel
	case "trace":
		Access = TraceLevel
	default:
		Access = InfoLevel
	}
}

func Error(loc string, data any) {
	write(ErrorLevel, "ERROR", loc, data)
}

func Warning(loc string, data any) {
	write(WarningLevel, "WARNING", loc, data)
}

func Info(loc string, data any) {
	write(InfoLevel, "INFO", loc, data)
}

func Debug(loc string, data any) {
	write(DebugLevel, "DEBUG", loc, data)
}

func Trace(loc string, data any) {
	write(TraceLevel, "TRACE", loc, data)
}

func write(level Level, name string, loc string, data any) {
	if Access < level {
		return
	}

	log.Printf("[%s] [%s] -> %v", name, loc, data)
}
