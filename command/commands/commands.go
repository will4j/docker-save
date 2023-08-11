package commands

import (
	"docker-thin/command/image"
	"docker-thin/docker"
	"github.com/spf13/cobra"
)

// AddCommands adds all the commands from cli/command to the root command
func AddCommands(cmd *cobra.Command, dockerCli docker.Cli) {
	cmd.AddCommand(
		image.NewSaveCommand(dockerCli),
	)
}
