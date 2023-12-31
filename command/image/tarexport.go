package image

import (
	"context"
	"docker-save/docker"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/moby/sys/symlink"
	"io"
	"os"
	"path/filepath"
	"strings"
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

type commonImageOptions struct {
	images    []string
	workdir   string
	keep      bool
	cacheFrom string
}

// ExportImages export images
func ExportImages(dockerCli docker.Cli, images []string) (io.ReadCloser, error) {
	// check docker service & image first
	err := imageInspectCheck(dockerCli, images)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return dockerCli.Client().ImageSave(ctx, images)
}

// GetPatternFunc is a function which used to generate temp dir pattern
type GetPatternFunc func() string

// ExportUntarImages export and untar images
func ExportUntarImages(dockerCli docker.Cli, opts commonImageOptions, getPatternFunc GetPatternFunc) (string, error) {
	if opts.cacheFrom != "" {
		// use cached untar dir
		return opts.cacheFrom, nil
	}

	untarDir, err := os.MkdirTemp(opts.workdir, getPatternFunc())
	if err != nil {
		return "", err
	}

	if err := doExportAndUntar(dockerCli, opts.images, untarDir); err != nil {
		return untarDir, err
	}
	return untarDir, nil
}

func ResolveManifests(workDir string) ([]manifestItem, error) {
	manifestPath, err := safePath(workDir, manifestFileName)
	if err != nil {
		return nil, err
	}
	manifestFile, err := os.Open(manifestPath)
	defer manifestFile.Close()

	var manifest []manifestItem
	if err := json.NewDecoder(manifestFile).Decode(&manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func doExportAndUntar(dockerCli docker.Cli, images []string, unTarDir string) error {
	imagesTar, err := ExportImages(dockerCli, images)
	if err != nil {
		return err
	}
	if err := archive.Untar(imagesTar, unTarDir, &archive.TarOptions{NoLchown: true}); err != nil {
		return err
	}
	return nil
}

func imageInspectCheck(dockerCli docker.Cli, images []string) error {
	_, err := ImageInspect(dockerCli, images)
	if err != nil {
		return err
	}
	return nil
}

func ImageInspect(dockerCli docker.Cli, images []string) ([]types.ImageInspect, error) {
	ctx := context.Background()
	getRefFunc := func(ref string) (types.ImageInspect, []byte, error) {
		return dockerCli.Client().ImageInspectWithRaw(ctx, ref)
	}
	resultArr := []types.ImageInspect{}
	for _, image := range images {
		inspect, _, err := getRefFunc(image)
		if err != nil {
			return nil, err
		}
		resultArr = append(resultArr, inspect)
	}
	return resultArr, nil
}

func safePath(base, path string) (string, error) {
	return symlink.FollowSymlinkInScope(filepath.Join(base, path), base)
}

func shouldCleanUntarDir(opts commonImageOptions) bool {
	if opts.keep {
		return false
	}
	if opts.cacheFrom != "" {
		return false
	}
	return true
}

func OmitString(str string, maxLength int) string {
	if len(str) <= maxLength {
		return str
	}
	tail := maxLength / 3
	head := maxLength - tail - 3
	return str[:head] + "..." + str[len(str)-tail:]
}

func ImagesConcatFmt(images []string) string {
	simplified := simplifyImageStr(images[0])
	if len(images) > 1 {
		simplified = simplified + "..."
	}
	return simplified
}

func simplifyImageStr(image string) string {
	if strings.Contains(image, "/") {
		tmp := strings.Split(image, "/")
		image = tmp[len(tmp)-1]
	}
	return strings.ReplaceAll(image, ":", "_")
}
