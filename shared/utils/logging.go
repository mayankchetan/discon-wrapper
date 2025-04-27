// Package utils provides shared utilities for both client and server
package utils

import (
	"log"
)

// GH-Cp gen: DebugLogger provides a standardized logging interface
// that respects debug levels for both client and server
type DebugLogger struct {
	DebugLevel int
	Prefix     string
}

// GH-Cp gen: NewDebugLogger creates a new DebugLogger with the specified debug level and prefix
func NewDebugLogger(debugLevel int, prefix string) *DebugLogger {
	return &DebugLogger{
		DebugLevel: debugLevel,
		Prefix:     prefix,
	}
}

// GH-Cp gen: LogAtLevel logs a message if the current debug level is at least the specified level
func (dl *DebugLogger) LogAtLevel(level int, format string, v ...interface{}) {
	if dl.DebugLevel >= level {
		if dl.Prefix != "" {
			format = dl.Prefix + ": " + format
		}
		log.Printf(format, v...)
	}
}

// GH-Cp gen: Debug logs at debug level 1 (basic info)
func (dl *DebugLogger) Debug(format string, v ...interface{}) {
	dl.LogAtLevel(1, format, v...)
}

// GH-Cp gen: Verbose logs at debug level 2 (verbose with payloads)
func (dl *DebugLogger) Verbose(format string, v ...interface{}) {
	dl.LogAtLevel(2, format, v...)
}

// GH-Cp gen: Error logs an error regardless of debug level
func (dl *DebugLogger) Error(format string, v ...interface{}) {
	if dl.Prefix != "" {
		format = dl.Prefix + ": ERROR: " + format
	} else {
		format = "ERROR: " + format
	}
	log.Printf(format, v...)
}