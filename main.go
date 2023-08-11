/*
Copyright © William Wang <williamw0825@gmail.com>
*/
package main

import (
	"docker-thin/command/commands"
	"docker-thin/docker"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func newDockerThinCommand(dockerCli *docker.DockerCli) *cobra.Command {

	cmd := &cobra.Command{
		Use:                   "docker-thin [OPTIONS] COMMAND [ARG...]",
		Short:                 "A self-sufficient runtime for containers",
		SilenceUsage:          true,
		SilenceErrors:         true,
		TraverseChildren:      true,
		DisableFlagsInUseLine: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   false,
			HiddenDefaultCmd:    true,
			DisableDescriptions: true,
		},
	}
	cmd.SetIn(dockerCli.In())
	cmd.SetOut(dockerCli.Out())
	cmd.SetErr(dockerCli.Err())

	commands.AddCommands(cmd, dockerCli)

	return cmd
}

func runDockerThin(dockerCli *docker.DockerCli) error {
	cmd := newDockerThinCommand(dockerCli)

	return cmd.Execute()
}

func main() {
	dockerCli := docker.NewDockerCli()
	logrus.SetOutput(dockerCli.Err())

	if err := runDockerThin(dockerCli); err != nil {
		if sterr, ok := err.(docker.StatusError); ok {
			if sterr.Status != "" {
				fmt.Fprintln(dockerCli.Err(), sterr.Status)
			}
			// StatusError should only be used for errors, and all errors should
			// have a non-zero exit status, so never exit with 0
			if sterr.StatusCode == 0 {
				os.Exit(1)
			}
			os.Exit(sterr.StatusCode)
		}
		fmt.Fprintln(dockerCli.Err(), err)
		os.Exit(1)
	}
}
