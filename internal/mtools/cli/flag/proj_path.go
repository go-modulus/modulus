package flag

import "github.com/urfave/cli/v2"

func ProjPathValue(ctx *cli.Context) string {
	return ctx.String("proj-path")
}
