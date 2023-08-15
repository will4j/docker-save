/*
Copyright © 2023 William Wang <williamw0825@gmail.com>
*/
package image

import (
	"docker-save/docker"
	"docker-save/docker/image"
	"fmt"
	"github.com/docker/go-units"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
	"time"
)

type statsOptions struct {
	commonImageOptions
}

// NewStatsCommand creates a new `docker-save stat` command
func NewStatsCommand(dockerCli docker.Cli) *cobra.Command {
	var opts statsOptions

	cmd := &cobra.Command{
		Use:   "stats [IMAGE...]",
		Short: "Stats image layers with command and size info",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.images = args
			return RunStats(dockerCli, opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.workdir, "workdir", "w", ".", "Directory for store tar files, default to current dir")
	flags.BoolVarP(&opts.keep, "keep", "k", false, "Keep workdir afterwards, default to auto clean")
	flags.StringVarP(&opts.cacheFrom, "cache-from", "c", "", "Use untar-images directory already exists other than export from docker")

	return cmd
}

// RunStats to stats image layers information
func RunStats(dockerCli docker.Cli, opts statsOptions) error {
	tempDirPattern := func() string {
		return "docker-stats-"
	}

	untarDir, err := ExportUntarImages(dockerCli, opts.commonImageOptions, tempDirPattern)
	if shouldCleanUntarDir(opts.commonImageOptions) && untarDir != "" {
		// must not be run before func outputSave
		defer os.RemoveAll(untarDir)
	}
	if err != nil {
		return err
	}

	manifests, err := ResolveManifests(untarDir)
	if err != nil {
		return err
	}

	for _, manifest := range manifests {
		configPath, err := safePath(untarDir, manifest.Config)
		if err != nil {
			return err
		}
		config, err := os.ReadFile(configPath)
		if err != nil {
			return err
		}

		img, err := image.NewFromJSON(config)
		diff_ids := img.RootFS.DiffIDs
		layers := manifest.Layers
		notEmptyHistory := filterNoEmptyHistory(img.History)

		if len(notEmptyHistory) != len(diff_ids) || len(notEmptyHistory) != len(layers) {
			return errors.New("NotEmptyLayers in history not equal to layers exists.")
		}

		printManifestStatsHead(dockerCli, manifest)
		for i, history := range notEmptyHistory {
			layerPath, _ := safePath(untarDir, layers[i])
			layerInfo, _ := os.Stat(layerPath)
			statsItem := LayerStatsItem{
				Number:  i + 1,
				DiffID:  diff_ids[i],
				Layer:   layers[i],
				Command: history.CreatedBy,
				Created: history.Created,
				Size:    layerInfo.Size(),
			}
			fmt.Fprintln(dockerCli.Out(), statsItem.Format())
		}
	}
	return nil
}

type LayerStatsItem struct {
	Number  int           `json:"number"`
	DiffID  digest.Digest `json:"diff_id"`
	Layer   string        `json:"layer"`
	Created *time.Time    `json:"created,omitempty"`
	Command string        `json:"command"`
	Size    int64         `json:"size"`
}

func (layer LayerStatsItem) Format() string {
	return fmt.Sprintf("Layer %2d: Size %8s, %-64s DiffID: %s Layer: %s",
		layer.Number,
		units.HumanSizeWithPrecision(float64(layer.Size), 5),
		ts(layer.Command, 64),
		layer.DiffID,
		layer.Layer)
}

var regex = regexp.MustCompile(`\s+`)

// truncate string
func ts(str string, length int) string {
	str = regex.ReplaceAllString(str, " ")
	str = strings.TrimSuffix(str, " # buildkit")
	if len(str) <= length {
		return str
	}
	tail := length / 3
	head := length - tail - 3
	return str[:head] + "..." + str[len(str)-tail:]
}

func filterNoEmptyHistory(history []image.History) []image.History {
	tmp := history[:0]
	for _, h := range history {
		if !h.EmptyLayer {
			tmp = append(tmp, h)
		}
	}
	return tmp
}

func printManifestStatsHead(dockerCli docker.Cli, manifest manifestItem) {
	identity := ""
	if len(manifest.RepoTags) > 0 {
		identity = fmt.Sprintf("Image Tag: %s", manifest.RepoTags[0])
	} else {
		identity = fmt.Sprintf("Image Id: %s", manifest.Config[:12])
	}
	fmt.Fprintf(dockerCli.Out(), "Start Stats of %s\n", identity)
}
