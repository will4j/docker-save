package client

import (
	cmd "docker-save/docker/utils"
	"io"
)

// ImageSave retrieves one or more images from the docker host as an io.ReadCloser.
// It's up to the caller to store the images and close the stream.
func (cli *Client) ImageSave(imageIDs []string) (io.ReadCloser, error) {
	subCmd := append([]string{"save"}, imageIDs...)
	stdout, stderr, err := cmd.OutputPipe("docker", subCmd...)
	defer stderr.Close()
	return stdout, err
}
