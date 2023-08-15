package image

import (
	"encoding/json"
	"errors"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Image stores partial docker image configuration
type Image struct {
	// RootFS contains information about the image's RootFS, including the
	// layer IDs.
	RootFS  *RootFS   `json:"rootfs,omitempty"`
	History []History `json:"history,omitempty"`

	// rawJSON caches the immutable JSON associated with this image.
	rawJSON []byte
}

// RootFS describes images root filesystem
// This is currently a placeholder that only supports layers. In the future
// this can be made into an interface that supports different implementations.
type RootFS struct {
	Type    string          `json:"type"`
	DiffIDs []digest.Digest `json:"diff_ids,omitempty"`
}

// History stores build commands that were used to create an image
type History = ocispec.History

// NewFromJSON creates an Image configuration from json.
func NewFromJSON(src []byte) (*Image, error) {
	img := &Image{}

	if err := json.Unmarshal(src, img); err != nil {
		return nil, err
	}
	if img.RootFS == nil {
		return nil, errors.New("invalid image JSON, no RootFS key")
	}

	img.rawJSON = src

	return img, nil
}
