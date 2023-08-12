/*
Copyright © 2023 William Wang <williamw0825@gmail.com>
*/
package image

import (
	"docker-save/docker"
	"io"

	"github.com/docker/cli/cli/command"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type saveOptions struct {
	images []string
	output string
}

// NewSaveCommand creates a new `docker save` command
func NewSaveCommand(dockerCli docker.Cli) *cobra.Command {
	var opts saveOptions

	cmd := &cobra.Command{
		Use:   "save [OPTIONS] IMAGE [IMAGE...]",
		Short: "Save one or more images to a tar archive (streamed to STDOUT by default)",
		Args:  docker.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.images = args
			return RunSave(dockerCli, opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.output, "output", "o", "", "Write to a file, instead of STDOUT")

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

	responseBody, err := dockerCli.Client().ImageSave(opts.images)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if opts.output == "" {
		_, err := io.Copy(dockerCli.Out(), responseBody)
		return err
	}

	return command.CopyToFile(opts.output, responseBody)
}