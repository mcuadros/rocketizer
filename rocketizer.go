package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mcuadros/rocketizer/command"
)

func main() {
	parser := flags.NewNamedParser("test", flags.Default)
	parser.AddCommand(
		"convert",
		"Convert a Dockerfile into a ACI container",
		"",
		new(command.Convert),
	)

	_, err := parser.Parse()
	if err != nil {
		if _, ok := err.(*flags.Error); ok {
			parser.WriteHelp(os.Stdout)
		}

		os.Exit(1)
	}
}
