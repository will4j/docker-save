package docker

import (
	"fmt"
	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/client"
	"github.com/moby/term"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"strings"
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
	in      *streams.In
	out     *streams.Out
	err     io.Writer
	init    sync.Once
	initErr error
	client  client.APIClient
}

func (cli *DockerCli) initialize() error {
	cli.init.Do(func() {
		if cli.initErr != nil {
			cli.initErr = errors.Wrap(cli.initErr, "unable to resolve docker endpoint")
			return
		}
		if cli.client == nil {
			if cli.client, cli.initErr = newDockerAPIClient(); cli.initErr != nil {
				return
			}
		}
	})
	return nil
}

func newDockerAPIClient() (client.APIClient, error) {
	host := os.Getenv("DOCKER_HOST")
	var clientOpts []client.Opt

	switch strings.Split(host, ":")[0] {
	case "ssh":
		helper, err := connhelper.GetConnectionHelper(host)
		if err != nil {
			fmt.Println("docker host", err)
		}
		clientOpts = append(clientOpts, func(c *client.Client) error {
			httpClient := &http.Client{
				Transport: &http.Transport{
					DialContext: helper.Dialer,
				},
			}
			return client.WithHTTPClient(httpClient)(c)
		})
		clientOpts = append(clientOpts, client.WithHost(helper.Host))
		clientOpts = append(clientOpts, client.WithDialContext(helper.Dialer))

	default:

		if os.Getenv("DOCKER_TLS_VERIFY") != "" && os.Getenv("DOCKER_CERT_PATH") == "" {
			os.Setenv("DOCKER_CERT_PATH", "~/.docker")
		}

		clientOpts = append(clientOpts, client.FromEnv)
	}

	clientOpts = append(clientOpts, client.WithAPIVersionNegotiation())
	dockerClient, err := client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, err
	}
	return dockerClient, nil
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
