package logger

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// LogLevel represents logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger represents the application logger
type Logger struct {
	level  LogLevel
	format string
	output *log.Logger
}

var globalLogger *Logger

// Init initializes the global logger
func Init() error {
	level := parseLogLevel(viper.GetString("log.level"))
	format := viper.GetString("log.format")
	
	logger := &Logger{
		level:  level,
		format: format,
		output: log.New(os.Stderr, "", log.LstdFlags),
	}
	
	globalLogger = logger
	return nil
}

// parseLogLevel parses string log level to LogLevel
func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	if globalLogger != nil && globalLogger.level <= DEBUG {
		globalLogger.log("DEBUG", msg, args...)
	}
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	if globalLogger != nil && globalLogger.level <= INFO {
		globalLogger.log("INFO", msg, args...)
	}
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	if globalLogger != nil && globalLogger.level <= WARN {
		globalLogger.log("WARN", msg, args...)
	}
}

// Error logs an error message
func Error(msg string, args ...interface{}) {
	if globalLogger != nil && globalLogger.level <= ERROR {
		globalLogger.log("ERROR", msg, args...)
	}
}

// log formats and outputs a log message
func (l *Logger) log(level, msg string, args ...interface{}) {
	formattedMsg := fmt.Sprintf(msg, args...)
	
	switch l.format {
	case "json":
		l.output.Printf(`{"level":"%s","message":"%s"}`, level, formattedMsg)
	default:
		l.output.Printf("[%s] %s", level, formattedMsg)
	}
}

// SetLevel sets the logging level
func SetLevel(level LogLevel) {
	if globalLogger != nil {
		globalLogger.level = level
	}
}