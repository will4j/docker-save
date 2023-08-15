package image

import (
	"context"
	"docker-save/docker"
	"github.com/docker/docker/pkg/archive"
	"io"
)

const (
	manifestFileName           = "manifest.json"
	legacyLayerFileName        = "layer.tar"
	legacyConfigFileName       = "json"
	legacyVersionFileName      = "VERSION"
	legacyRepositoriesFileName = "repositories"
)

type manifestItem struct {
	Config   string
	RepoTags []string
	Layers   []string
}

// ExportImages export images
func ExportImages(dockerCli docker.Cli, images []string) (io.ReadCloser, error) {
	// check docker service & image first
	ctx := context.Background()
	err := imageInspectCheck(ctx, dockerCli, images)
	if err != nil {
		return nil, err
	}

	return dockerCli.Client().ImageSave(ctx, images)
}

// ExportUntarImages export and untar images
func ExportUntarImages(dockerCli docker.Cli, images []string, unTarDir string) error {
	imagesTar, err := ExportImages(dockerCli, images)
	if err != nil {
		return err
	}
	if err := archive.Untar(imagesTar, unTarDir, &archive.TarOptions{NoLchown: true}); err != nil {
		return err
	}
	return nil
}

func imageInspectCheck(ctx context.Context, dockerCli docker.Cli, images []string) error {
	client := dockerCli.Client()
	getRefFunc := func(ref string) (interface{}, []byte, error) {
		return client.ImageInspectWithRaw(ctx, ref)
	}
	for _, image := range images {
		_, _, err := getRefFunc(image)
		if err != nil {
			return err
		}
	}
	return nil
}
