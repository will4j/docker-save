package client

import (
	cmd "docker-save/docker/utils"
	"errors"
	"io"
)

// ImageInspect inspect one or more images from the docker host as an io.ReadCloser.
// It's up to the caller to store the images info and close the stream.
func (cli *Client) ImageInspect(imageIDs []string) (io.ReadCloser, error) {
	subCmd := append([]string{"image", "inspect"}, imageIDs...)
	stdout, stderr, err := cmd.OutputPipe("docker", subCmd...)
	if err != nil {
		return nil, err
	}
	defer stderr.Close()

	msg, err := io.ReadAll(stderr)
	if len(msg) != 0 {
		return nil, errors.New(string(msg))
	}
	return stdout, err
}
