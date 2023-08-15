/*
Copyright © 2023 William Wang <williamw0825@gmail.com>
*/
package image

import (
	"docker-save/docker"
	"encoding/json"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/pkg/archive"
	"github.com/moby/sys/symlink"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
)

type saveOptions struct {
	images  []string
	output  string
	workdir string
	keep    bool
	last    int
	latest  bool
}

// NewSaveCommand creates a new `docker save` command
func NewSaveCommand(dockerCli docker.Cli) *cobra.Command {
	var opts saveOptions

	cmd := &cobra.Command{
		Use: "docker-save IMAGE [IMAGE...]",
		Long: `A tool for saving docker images to a tar archive (streamed to STDOUT by default)
add support for filtering image layers`,
		Args: docker.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.images = args
			return RunSave(dockerCli, opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.output, "output", "o", "", "Write to a file, instead of STDOUT")
	flags.StringVarP(&opts.workdir, "workdir", "w", ".", "Directory for store tar files, default to current dir")
	flags.IntVarP(&opts.last, "last", "l", 0, "Export the last n image layers for each image")
	flags.BoolVarP(&opts.latest, "latest", "L", false, "Only export the latest image layer for each image")
	flags.BoolVarP(&opts.keep, "keep", "k", false, "Keep workdir afterwards, default to auto clean")

	return cmd
}

// RunSave performs a save against the engine based on the specified options
func RunSave(dockerCli docker.Cli, opts saveOptions) error {
	if opts.output == "" && dockerCli.Out().IsTerminal() {
		return errors.New("cowardly refusing to save to a terminal. Use the -o flag or redirect")
	}

	if err := command.ValidateOutputPath(opts.output); err != nil {
		return errors.Wrap(err, "failed to save image")
	}

	if needToFilterImageLayers(opts) {
		return exportImagesWithFilter(dockerCli, opts)
	} else {
		imagesTar, err := ExportImages(dockerCli, opts.images)
		if err != nil {
			return err
		}
		return outputSave(dockerCli, opts.output, imagesTar)
	}
}

func exportImagesWithFilter(dockerCli docker.Cli, opts saveOptions) error {
	workDir, err := os.MkdirTemp(opts.workdir, "docker-save-")
	if err != nil {
		return err
	}
	if !opts.keep && workDir != "" {
		defer os.RemoveAll(workDir)
	}

	if err := ExportUntarImages(dockerCli, opts.images, workDir); err != nil {
		return err
	}

	manifestPath, err := safePath(workDir, manifestFileName)
	if err != nil {
		return err
	}
	manifestFile, err := os.Open(manifestPath)
	defer manifestFile.Close()

	var manifest []manifestItem
	if err := json.NewDecoder(manifestFile).Decode(&manifest); err != nil {
		return err
	}

	excludedLayers := []string{}
	for _, m := range manifest {
		layers := m.Layers
		excludedLayers = append(excludedLayers, layersToExclude(layers, opts)...)
	}
	tarOptions := &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: excludedLayers,
	}
	tar, err := archive.TarWithOptions(workDir, tarOptions)
	if err != nil {
		return err
	}

	return outputSave(dockerCli, opts.output, tar)
}

func outputSave(dockerCli docker.Cli, output string, body io.ReadCloser) error {
	defer body.Close()
	if output == "" {
		_, err := io.Copy(dockerCli.Out(), body)
		return err
	}

	return command.CopyToFile(output, body)
}

func safePath(base, path string) (string, error) {
	return symlink.FollowSymlinkInScope(filepath.Join(base, path), base)
}

func needToFilterImageLayers(opts saveOptions) bool {
	if opts.last > 0 {
		return true
	}
	if opts.latest {
		return true
	}
	return false
}

func layersToExclude(layers []string, opts saveOptions) []string {
	end := len(layers)
	if opts.latest {
		end = end - 1
	} else if opts.last > 0 {
		end = end - opts.last
	}
	if end < 1 {
		return []string{}
	}
	return layers[:end]
}
