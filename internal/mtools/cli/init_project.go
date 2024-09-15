package cli

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log/slog"
)

type InitProject struct {
	logger *slog.Logger
}

func NewInitProject(
	logger *slog.Logger,
) *InitProject {
	return &InitProject{
		logger: logger,
	}
}

func NewCommand(c *InitProject) *cli.Command {
	return &cli.Command{
		Name: "init",
		Usage: `Inits a project with the base Modulus structure.
	Uses interactive prompts to create the project.
	Example: ./bin/console init
`,
		Action: c.Invoke,
	}
}

func (c *InitProject) Invoke(
	ctx *cli.Context,
) error {

	fmt.Println("Start initializing a project")

	fmt.Println(
		"Congratulations! Your project has been initialized. Please, add your first module.",
	)

	return nil
}
