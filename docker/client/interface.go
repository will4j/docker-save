package client

import (
	"io"
)

// APIClient defines API client methods
type APIClient interface {
	ImageSave(images []string) (io.ReadCloser, error)
	ImageInspect(images []string) (io.ReadCloser, error)
}

// Client is the API client that performs all operations
// against a docker server.
type Client struct {
}
