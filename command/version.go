package command

import (
	"fmt"

	"github.com/str1ngs/ansi/color"
)

type Version struct {
	Version string
	Build   string
}

func (v *Version) Execute(args []string) error {
	fmt.Printf("rocketizer %s - build %s\n",
		color.Green(v.Version),
		color.Green(v.Build),
	)

	return nil
}
