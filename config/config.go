// Package config provides configuration management for the SNET panel.
package config

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed version
var version string

//go:embed name
var name string

type LogLevel string

const (
	Debug   LogLevel = "debug"
	Info    LogLevel = "info"
	Notice  LogLevel = "notice"
	Warning LogLevel = "warning"
	Error   LogLevel = "error"
)

func GetVersion() string {
	return strings.TrimSpace(version)
}

func GetName() string {
	return strings.TrimSpace(name)
}

func GetLogLevel() LogLevel {
	if IsDebug() {
		return Debug
	}
	logLevel := os.Getenv("SNET_LOG_LEVEL")
	if logLevel == "" {
		return Info
	}
	return LogLevel(logLevel)
}

func IsDebug() bool {
	return os.Getenv("SNET_DEBUG") == "true"
}

func GetBinFolderPath() string {
	p := os.Getenv("SNET_BIN_FOLDER")
	if p == "" {
		p = "bin"
	}
	return p
}

func getBaseDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	exeDir := filepath.Dir(exePath)
	exeDirLower := strings.ToLower(filepath.ToSlash(exeDir))
	if strings.Contains(exeDirLower, "/appdata/local/temp/") || strings.Contains(exeDirLower, "/go-build") {
		wd, err := os.Getwd()
		if err != nil {
			return "."
		}
		return wd
	}
	return exeDir
}

func GetDBFolderPath() string {
	p := os.Getenv("SNET_DB_FOLDER")
	if p != "" {
		return p
	}
	if runtime.GOOS == "windows" {
		return getBaseDir()
	}
	return "/etc/snet"
}

func GetDBPath() string {
	return fmt.Sprintf("%s/%s.db", GetDBFolderPath(), GetName())
}

func GetLogFolder() string {
	p := os.Getenv("SNET_LOG_FOLDER")
	if p != "" {
		return p
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(".", "log")
	}
	return "/var/log/snet"
}

func GetVPNConfigPath() string {
	p := os.Getenv("SNET_VPN_CONFIG")
	if p != "" {
		return p
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(getBaseDir(), "vpn-configs")
	}
	return "/etc/snet/vpn"
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}
