package db

import "github.com/urfave/cli/v2"

func NewDbCommand(updateSqlc *UpdateSQLCConfig) *cli.Command {
	return &cli.Command{
		Name: "db",
		Usage: `A set of commands for working with PostgreSQL database in modules.
Example: mtools db
`,
		Subcommands: []*cli.Command{
			NewUpdateSQLCConfigCommand(updateSqlc),
		},
	}
}
