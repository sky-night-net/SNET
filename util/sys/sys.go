package sys

import (
	"bytes"
	"os/exec"
)

func RunCommand(name string, arg ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(name, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func Execute(cmd string) (string, error) {
	stdout, stderr, err := RunCommand("bash", "-c", cmd)
	if err != nil {
		if stderr != "" {
			return "", fmt.Errorf("%v: %s", err, stderr)
		}
		return "", err
	}
	return stdout, nil
}
