package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mcuadros/rocketizer/command"
)

var version string
var build string

func main() {
	parser := flags.NewNamedParser("rocketizer", flags.Default)
	parser.AddCommand(
		"version",
		"Version and build information",
		"",
		&command.Version{Version: version, Build: build},
	)

	parser.AddCommand(
		"convert",
		"Convert a Dockerfile into a ACI container",
		"",
		&command.Convert{},
	)

	_, err := parser.Parse()
	if err != nil {
		if _, ok := err.(*flags.Error); ok {
			parser.WriteHelp(os.Stdout)
		}

		os.Exit(1)
	}
}
