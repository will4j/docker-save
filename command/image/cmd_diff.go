package image

import (
	"docker-save/docker"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/spf13/cobra"
)

type diffOptions struct {
	images []string
}

// NewDiffCommand compare two images and show diff between layers
func NewDiffCommand(dockerCli docker.Cli) *cobra.Command {
	var opts diffOptions

	cmd := &cobra.Command{
		Use:   "diff IMAGE IMAGE",
		Short: "Show layer difference between two images",
		Args:  docker.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.images = args
			return RunDiff(dockerCli, opts)
		},
	}
	return cmd
}

func RunDiff(dockerCli docker.Cli, opts diffOptions) error {
	inspects, err := ImageInspect(dockerCli, opts.images)
	if err != nil {
		return err
	}
	layers0 := inspects[0].RootFS.Layers
	layers1 := inspects[1].RootFS.Layers
	size0 := len(layers0)
	size1 := len(layers1)

	loop := size0
	if size0 < size1 {
		loop = size1
	}
	printDiffHead(dockerCli, inspects[0], inspects[1])
	diffCount := 0
	paramDiffCount := 0
	startDiff := false
	for i := 0; i < loop; i++ {
		layer0 := ""
		if i < size0 {
			layer0 = layers0[i]
		}
		layer1 := ""
		if i < size1 {
			layer1 = layers1[i]
		}
		if printDiffLayer(dockerCli, layer0, layer1) {
			diffCount += 1
			startDiff = true
		}
		if startDiff {
			paramDiffCount += 1
		}
	}
	printDiffCount(dockerCli, diffCount, paramDiffCount)
	return nil
}

func printDiffHead(dockerCli docker.Cli, inspect0 types.ImageInspect, inspect1 types.ImageInspect) {
	fmt.Fprintf(dockerCli.Out(), "%35s %35s\n", OmitString(inspect0.RepoTags[0], 35), OmitString(inspect1.RepoTags[0], 35))
	fmt.Fprintln(dockerCli.Out(), "")
}

func printDiffLayer(dockerCli docker.Cli, layer0 string, layer1 string) bool {
	if layer0 == layer1 {
		fmt.Fprintln(dockerCli.Out(), layer0)
		return false
	} else {
		fmt.Fprintf(dockerCli.Out(), "%35.35s %35.35s\n",
			OmitString(layer0, 35), OmitString(layer1, 35))
		return true
	}
}

func printDiffCount(dockerCli docker.Cli, diffCount int, paramDiffCount int) {
	fmt.Fprintln(dockerCli.Out(), "\nNumber of Different Layers:", diffCount)
	fmt.Fprintln(dockerCli.Out(), "\nParam of Export Different Layers:", paramDiffCount)
}
