package utils

import (
	"errors"
	"io"
	"os/exec"
)

func OutputPipe(name string, arg ...string) (io.ReadCloser, error) {
	cmd := exec.Command(name, arg...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer stderr.Close()
	msg, err := io.ReadAll(stderr)
	if len(msg) != 0 {
		return nil, errors.New(string(msg))
	}
	return stdout, nil
}
