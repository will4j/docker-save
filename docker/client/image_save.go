package client

import (
	cmd "docker-save/docker/utils"
	"io"
	"strings"
)

// ImageSave retrieves one or more images from the docker host as an io.ReadCloser.
// It's up to the caller to store the images and close the stream.
func (cli *Client) ImageSave(imageIDs []string) (io.ReadCloser, error) {
	return cmd.OutputPipe("docker", "save", strings.Join(imageIDs, " "))
}
