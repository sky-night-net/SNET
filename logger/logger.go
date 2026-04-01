// Package logger provides structured logging for SNET.
package logger

import (
	"os"
	"sync"

	"github.com/op/go-logging"
)

var logger *logging.Logger
var once sync.Once

func InitLogger(level logging.Level) {
	once.Do(func() {
		newLogger := logging.MustGetLogger("snet")

		backendStdout := logging.NewLogBackend(os.Stdout, "", 0)
		format := logging.MustStringFormatter(
			`%{time:2006/01/02 15:04:05} %{level:.4s} - %{message}`,
		)
		backendStdoutFormatter := logging.NewBackendFormatter(backendStdout, format)
		backendStdoutLeveled := logging.AddModuleLevel(backendStdoutFormatter)
		backendStdoutLeveled.SetLevel(level, "")
		newLogger.SetBackend(backendStdoutLeveled)

		logger = newLogger
	})
}

func getLogger() *logging.Logger {
	if logger == nil {
		InitLogger(logging.INFO)
	}
	return logger
}

func Debug(args ...interface{})                 { getLogger().Debug(args...) }
func Debugf(format string, args ...interface{}) { getLogger().Debugf(format, args...) }
func Info(args ...interface{})                  { getLogger().Info(args...) }
func Infof(format string, args ...interface{})  { getLogger().Infof(format, args...) }
func Warning(args ...interface{})               { getLogger().Warning(args...) }
func Warningf(format string, args ...interface{}) { getLogger().Warningf(format, args...) }
func Error(args ...interface{})                 { getLogger().Error(args...) }
func Errorf(format string, args ...interface{}) { getLogger().Errorf(format, args...) }

var logBuffer []string
var mu sync.RWMutex
const maxBufferSize = 5000

func AddLog(line string) {
	mu.Lock()
	defer mu.Unlock()
	logBuffer = append(logBuffer, line)
	if len(logBuffer) > maxBufferSize {
		logBuffer = logBuffer[len(logBuffer)-maxBufferSize:]
	}
}

func GetLogs(count int, level string) []string {
	mu.RLock()
	defer mu.RUnlock()
	if count <= 0 || count > len(logBuffer) {
		count = len(logBuffer)
	}
	start := len(logBuffer) - count
	if start < 0 {
		start = 0
	}
	result := make([]string, count)
	copy(result, logBuffer[start:])
	return result
}
