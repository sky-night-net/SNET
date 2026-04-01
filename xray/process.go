package xray

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/sky-night-net/snet/config"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/util/common"
)

func GetBinaryName() string {
	return fmt.Sprintf("xray-%s-%s", runtime.GOOS, runtime.GOARCH)
}

func GetBinaryPath() string {
	return config.GetBinFolderPath() + "/" + GetBinaryName()
}

func GetConfigPath() string {
	return config.GetBinFolderPath() + "/config.json"
}

func GetGeositePath() string {
	return config.GetBinFolderPath() + "/geosite.dat"
}

func GetGeoipPath() string {
	return config.GetBinFolderPath() + "/geoip.dat"
}

func stopProcess(p *Process) {
	p.Stop()
}

type Process struct {
	*process
}

func NewProcess(xrayConfig *Config) *Process {
	p := &Process{newProcess(xrayConfig)}
	runtime.SetFinalizer(p, stopProcess)
	return p
}

type process struct {
	cmd *exec.Cmd

	version string
	apiPort int

	onlineClients []string

	config     *Config
	configPath string
	logWriter  *LogWriter
	exitErr    error
	startTime  time.Time
}

func newProcess(config *Config) *process {
	return &process{
		version:   "Unknown",
		config:    config,
		logWriter: NewLogWriter(),
		startTime: time.Now(),
	}
}

func (p *process) IsRunning() bool {
	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}
	if p.cmd.ProcessState == nil {
		return true
	}
	return false
}

func (p *process) GetErr() error {
	return p.exitErr
}

func (p *process) GetResult() string {
	if len(p.logWriter.lastLine) == 0 && p.exitErr != nil {
		return p.exitErr.Error()
	}
	return p.logWriter.lastLine
}

func (p *process) GetVersion() string {
	return p.version
}

func (p *Process) GetAPIPort() int {
	return p.apiPort
}

func (p *Process) GetConfig() *Config {
	return p.config
}

func (p *Process) GetUptime() uint64 {
	return uint64(time.Since(p.startTime).Seconds())
}

func (p *process) refreshAPIPort() {
	for _, inbound := range p.config.InboundConfigs {
		if inbound.Tag == "api" {
			p.apiPort = inbound.Port
			break
		}
	}
}

func (p *process) refreshVersion() {
	cmd := exec.Command(GetBinaryPath(), "-version")
	data, err := cmd.Output()
	if err != nil {
		p.version = "Unknown"
	} else {
		datas := bytes.Split(data, []byte(" "))
		if len(datas) <= 1 {
			p.version = "Unknown"
		} else {
			p.version = string(datas[1])
		}
	}
}

func (p *process) Start() (err error) {
	if p.IsRunning() {
		return errors.New("xray is already running")
	}

	defer func() {
		if err != nil {
			logger.Error("Failure in running xray-core process: ", err)
			p.exitErr = err
		}
	}()

	data, err := json.MarshalIndent(p.config, "", "  ")
	if err != nil {
		return common.NewErrorf("Failed to generate XRAY configuration files: %v", err)
	}

	err = os.MkdirAll(config.GetLogFolder(), 0o770)
	if err != nil {
		logger.Warningf("Failed to create log folder: %s", err)
	}

	configPath := GetConfigPath()
	if p.configPath != "" {
		configPath = p.configPath
	}
	err = os.WriteFile(configPath, data, fs.ModePerm)
	if err != nil {
		return common.NewErrorf("Failed to write configuration file: %v", err)
	}

	cmd := exec.Command(GetBinaryPath(), "-c", configPath)
	p.cmd = cmd

	cmd.Stdout = p.logWriter
	cmd.Stderr = p.logWriter

	go func() {
		err := cmd.Run()
		if err != nil {
			if runtime.GOOS == "windows" {
				errStr := strings.ToLower(err.Error())
				if strings.Contains(errStr, "exit status 1") {
					p.exitErr = err
					return
				}
			}
			logger.Error("Failure in running xray-core:", err)
			p.exitErr = err
		}
	}()

	p.refreshVersion()
	p.refreshAPIPort()

	return nil
}

func (p *process) Stop() error {
	if !p.IsRunning() {
		return errors.New("xray is not running")
	}

	if runtime.GOOS == "windows" {
		return p.cmd.Process.Kill()
	} else {
		return p.cmd.Process.Signal(syscall.SIGTERM)
	}
}

func writeCrashReport(m []byte) error {
	crashReportPath := config.GetBinFolderPath() + "/core_crash_" + time.Now().Format("20060102_150405") + ".log"
	return os.WriteFile(crashReportPath, m, os.ModePerm)
}
