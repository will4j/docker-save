package client

import (
	"io"
	"log"
	"os/exec"
	"strings"
)

// ImageSave retrieves one or more images from the docker host as an io.ReadCloser.
// It's up to the caller to store the images and close the stream.
func (cli *Client) ImageSave(imageIDs []string) (io.ReadCloser, error) {
	cmd := exec.Command("docker", "save", strings.Join(imageIDs, " "))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return stdout, nil
}
