package utils

import (
	"io"
	"os/exec"
)

func OutputPipe(name string, arg ...string) (io.ReadCloser, io.ReadCloser, error) {
	cmd := exec.Command(name, arg...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return stdout, nil, err
	}
	if err := cmd.Start(); err != nil {
		return stdout, stderr, err
	}
	return stdout, stderr, nil
}
