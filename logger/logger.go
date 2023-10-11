// Copyright 2016 Viacheslav Chimishuk <vchimishuk@yandex.ru>
//
// This file is part of Chub.
//
// Chub is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Chub is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Chub. If not, see <http://www.gnu.org/licenses/>.

// logger package implement simple stderr logging functionality.
// Actually it is a wrapper around standard log package which
// adds log levels support.
package logger

import (
	"fmt"
	"log"
	"os"
)

// Available log levels.
const (
	LevelDebug = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
	LevelNone
)

// Default logging flags combination to use.
const defaultFlags = log.Ltime | log.Lmicroseconds

// String representations on of log levels.
var levelNames = map[int]string{
	LevelDebug:   "DEBUG",
	LevelInfo:    "INFO",
	LevelWarning: "WARNING",
	LevelError:   "ERROR",
	LevelFatal:   "FATAL",
}

// The logger structure.
type logger struct {
	level     int
	stdLogger *log.Logger
}

// Default logger object which is used for logging by all
// public functions.
var defaultLogger *logger = newLogger(LevelDebug)

// newLogger creates new logger object.
func newLogger(level int) *logger {
	stdLogger := log.New(os.Stderr, "", defaultFlags)

	return &logger{level: level, stdLogger: stdLogger}
}

// output prints a message using standard logger from log package.
func (l *logger) output(level int, format string, args ...interface{}) {
	l.stdLogger.Output(3, fmt.Sprintf(format, args...))
}

// Level returns current log level used by the standard logger.
func Level() int {
	return defaultLogger.level
}

// SetLevel sets new logging level for the default logger.
func SetLevel(level int) {
	defaultLogger.level = level
}

// Debug logs debug leveled message.
func Debug(format string, args ...interface{}) {
	if defaultLogger.level <= LevelDebug {
		defaultLogger.output(LevelDebug, format, args...)
	}
}

// Info logs info leveled message.
func Info(format string, args ...interface{}) {
	if defaultLogger.level <= LevelInfo {
		defaultLogger.output(LevelInfo, format, args...)
	}
}

// Warning logs warning leveled message.
func Warning(format string, args ...interface{}) {
	if defaultLogger.level <= LevelWarning {
		defaultLogger.output(LevelWarning, format, args...)
	}
}

// Error logs error leveled message.
func Error(format string, args ...interface{}) {
	if defaultLogger.level <= LevelError {
		defaultLogger.output(LevelError, format, args...)
	}
}

// Fatal logs fatal leveled message and panics.
func Fatal(format string, args ...interface{}) {
	if defaultLogger.level <= LevelFatal {
		defaultLogger.output(LevelFatal, format, args...)
		panic(fmt.Sprintf(format, args...))
	}
}
