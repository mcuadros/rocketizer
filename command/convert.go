package command

import (
	"os"

	"github.com/docker/docker/builder/parser"
	"github.com/mcuadros/rocketizer/transformer"
)

type Convert struct {
	Dockerfile string `short:"d" long:"dockerfile" description:"Dockerfile to be converted." required:"true"`
	Name       string `short:"n" long:"name" description:"human readable index to the container image." required:"true"`
	Version    string `short:"v" long:"version" description:"container name." required:"true"`
}

func (c *Convert) Execute(args []string) error {
	t, err := transformer.NewToRocket(c.Name, c.Version, "linux", "amd64")
	if err != nil {
		return err
	}

	n, err := c.getNodesFromDockerfile()
	if err != nil {
		return err
	}

	t.Process(n)
	t.Print()

	return nil
}

func (c *Convert) getNodesFromDockerfile() (*parser.Node, error) {
	f, err := os.Open(c.Dockerfile)
	if err != nil {
		return nil, err
	}

	n, err := parser.Parse(f)
	if err != nil {
		return nil, err
	}

	return n, nil
}
