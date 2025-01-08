package flag

import (
	"github.com/urfave/cli/v2"
)

func NewSilent(usage string) cli.Flag {
	return &cli.BoolFlag{
		Name:    "silent",
		Usage:   usage,
		Aliases: []string{"s"},
	}
}

func SilentValue(ctx *cli.Context) bool {
	return ctx.Bool("silent")
}
