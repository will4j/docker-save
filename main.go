/*
Copyright © William Wang <williamw0825@gmail.com>
*/
package main

import (
	"docker-save/command/commands"
	"docker-save/command/image"
	"docker-save/docker"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func newDockerSaveCommand(dockerCli *docker.DockerCli) *cobra.Command {

	rootCmd := image.NewSaveCommand(dockerCli)

	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.TraverseChildren = true

	rootCmd.SetIn(dockerCli.In())
	rootCmd.SetOut(dockerCli.Out())
	rootCmd.SetErr(dockerCli.Err())

	commands.AddCommands(rootCmd, dockerCli)

	return rootCmd
}

func runDockerSave(dockerCli *docker.DockerCli) error {
	rootCmd := newDockerSaveCommand(dockerCli)
	return rootCmd.Execute()
}

func main() {
	dockerCli := docker.NewDockerCli()
	logrus.SetOutput(dockerCli.Err())

	if err := runDockerSave(dockerCli); err != nil {
		fmt.Fprintln(dockerCli.Err(), err)
		os.Exit(1)
	}
}
