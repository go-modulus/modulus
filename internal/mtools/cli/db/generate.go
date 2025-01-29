package db

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/internal/mtools/action"
	"github.com/go-modulus/modulus/internal/mtools/utils"
	"github.com/go-modulus/modulus/module"
	"github.com/urfave/cli/v2"
	"os/exec"
)

type Generate struct {
	action *action.UpdateSqlcConfig
}

func NewGenerate(action *action.UpdateSqlcConfig) *Generate {
	return &Generate{
		action: action,
	}
}

func NewGenerateCommand(updateSqlc *Generate) *cli.Command {
	return &cli.Command{
		Name: "generate",
		Usage: `Generates DTO and DAO files to work with DB. It uses SQLc compiler to do this action.
Example: mtools db generate
`,
		Action: updateSqlc.Invoke,
	}
}

func (c *Generate) Invoke(ctx *cli.Context) error {
	projPath := ctx.String("proj-path")
	manifest, err := module.LoadLocalManifest(projPath)
	if err != nil {
		fmt.Println(color.RedString("Cannot load the project manifest %s/modules.json: %s", projPath, err.Error()))
		return err
	}
	for _, md := range manifest.Modules {
		if !md.IsLocalModule {
			continue
		}
		storagePath := md.StoragePath(projPath)
		sqlcFile := storagePath + "/sqlc.yaml"
		if !utils.FileExists(sqlcFile) {
			continue
		}
		fmt.Println("Generate DTO and DAO files for", color.BlueString(md.Name))
		fmt.Println(fmt.Sprintf("Running %s ...", color.BlueString("sqlc -f "+sqlcFile+" generate")))
		cmd := exec.CommandContext(ctx.Context, "sqlc", "-f", sqlcFile, "generate")
		//stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		err = cmd.Start()
		if err != nil {
			fmt.Println(color.RedString("Cannot start the sqlc command: %s", err.Error()))
			var buf = make([]byte, 1024)
			_, _ = stderr.Read(buf)
			fmt.Println(color.RedString("Execution error:", string(buf)))
			return err
		}

		fmt.Println(
			color.GreenString("Generated successfully"),
		)
	}
	return nil
}
