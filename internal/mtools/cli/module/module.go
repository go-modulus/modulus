package module

import "github.com/urfave/cli/v2"

func NewModuleCommand(
	create *Create,
	install *Install,
	addCli *AddCli,
	addJsonApi *AddJsonApi,
) *cli.Command {
	return &cli.Command{
		Name: "module",
		Usage: `A set of commands for modules manipulations.
Example: mtools module
`,
		Subcommands: []*cli.Command{
			NewCreateCommand(create),
			NewInstallCommand(install),
			NewAddCliCommand(addCli),
			NewAddJsonApiCommand(addJsonApi),
		},
	}
}
