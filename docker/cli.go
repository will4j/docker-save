package docker

import (
	"docker-thin/docker/client"
	"fmt"
	"github.com/docker/cli/cli/streams"
	"github.com/moby/term"
	"io"
	"os"
	"sync"
)

// Streams is an interface which exposes the standard input and output streams
type Streams interface {
	In() *streams.In
	Out() *streams.Out
	Err() io.Writer
}

// Cli represents the docker command line client.
type Cli interface {
	Client() client.APIClient
	Streams
	SetIn(in *streams.In)
}

// DockerCli is an instance the docker command line client.
// Instances of the client can be returned from NewDockerCli.
type DockerCli struct {
	in     *streams.In
	out    *streams.Out
	err    io.Writer
	init   sync.Once
	client client.APIClient
}

func (cli *DockerCli) initialize() error {
	cli.init.Do(func() {
		if cli.client == nil {
			cli.client = &client.Client{}
		}
	})
	return nil
}

// Client returns the APIClient
func (cli *DockerCli) Client() client.APIClient {
	if err := cli.initialize(); err != nil {
		_, _ = fmt.Fprintf(cli.Err(), "Failed to initialize: %s\n", err)
		os.Exit(1)
	}
	return cli.client
}

// Out returns the writer used for stdout
func (cli *DockerCli) Out() *streams.Out {
	return cli.out
}

// Err returns the writer used for stderr
func (cli *DockerCli) Err() io.Writer {
	return cli.err
}

// SetIn sets the reader used for stdin
func (cli *DockerCli) SetIn(in *streams.In) {
	cli.in = in
}

// In returns the reader used for stdin
func (cli *DockerCli) In() *streams.In {
	return cli.in
}

// NewDockerCli returns a DockerCli instance with all operators applied on it.
// It applies by default the standard streams, and the content trust from
// environment.
func NewDockerCli() *DockerCli {
	cli := &DockerCli{}
	stdin, stdout, stderr := term.StdStreams()
	cli.in = streams.NewIn(stdin)
	cli.out = streams.NewOut(stdout)
	cli.err = stderr
	return cli
}
