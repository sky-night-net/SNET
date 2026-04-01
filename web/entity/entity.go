package entity

import "github.com/sky-night-net/snet/database/model"

type Msg struct {
	Success bool        `json:"success"`
	Msg     string      `json:"msg"`
	Obj     interface{} `json:"obj,omitempty"`
}

type AllSetting struct {
	WebPort      int    `json:"webPort"`
	WebBasePath  string `json:"webBasePath"`
	WebCertFile  string `json:"webCertFile"`
	WebKeyFile   string `json:"webKeyFile"`
	XrayBinPath  string `json:"xrayBinPath"`
	// Add other settings as needed for the frontend
}
