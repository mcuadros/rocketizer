package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/builder/parser"
	"github.com/mcuadros/rocketizer/transformer"
	"github.com/str1ngs/ansi/color"
)

type Convert struct {
	Dockerfile string `short:"d" long:"dockerfile" description:"Dockerfile to be converted."`
	Name       string `short:"n" long:"name" description:"human readable index to the container image." required:"true"`
	Version    string `short:"v" long:"version" description:"container name." required:"true"`
	Output     string `short:"o" long:"output" description:"output container name."`
}

func (c *Convert) Execute(args []string) error {
	if c.Dockerfile == "" {
		c.Dockerfile = "Dockerfile"
	}

	fmt.Printf("Building %s<%s>\n",
		color.Cyan(c.Name), color.Cyan(c.Version),
	)

	t, err := transformer.NewToRocket(c.Name, c.Version, "linux", "amd64")
	if err != nil {
		return err
	}

	var filename string
	if filename, err = c.getFilename(); err != nil {
		return err
	}

	t.BasePath = filepath.Dir(filename)

	fmt.Printf("Parsing Dockerfile %q\n", color.Cyan(filename))
	n, err := c.getNodesFromDockerfile(filename)
	if err != nil {
		return err
	}

	fmt.Printf("Compressing files... ")
	if err = t.Process(n); err != nil {
		return err
	}

	var file string
	if file, err = t.SaveToFile(c.Output); err != nil {
		return err
	}

	fmt.Println(color.Green("OK"))
	fmt.Printf("New ACI created %q\n", color.Green(file))

	return nil
}

func (c *Convert) getFilename() (string, error) {
	fInfo, err := os.Lstat(c.Dockerfile)
	if err != nil {
		return "", err
	}

	if fInfo.IsDir() {
		return filepath.Join(c.Dockerfile, "Dockerfile"), nil
	}

	return c.Dockerfile, nil
}

func (c *Convert) getNodesFromDockerfile(filename string) (*parser.Node, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	n, err := parser.Parse(f)
	if err != nil {
		return nil, err
	}

	return n, nil
}
