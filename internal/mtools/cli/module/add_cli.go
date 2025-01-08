package module

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/internal/mtools/cli/flag"
	"github.com/go-modulus/modulus/internal/mtools/files"
	"github.com/go-modulus/modulus/internal/mtools/utils"
	"github.com/go-modulus/modulus/module"
	"github.com/iancoleman/strcase"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"regexp"
)

var nameRegEx = regexp.MustCompile(`^[a-z0-9-]+$`)

type AddCliTmplVars struct {
	StructName  string
	CommandName string
}

type AddCli struct {
}

func NewAddCli() *AddCli {
	return &AddCli{}
}
func NewAddCliCommand(addCli *AddCli) *cli.Command {
	return &cli.Command{
		Name: "add-cli",
		Usage: `Add a boilerplate of the CLI command to the selected module.
Example: mtools module add-cli
Example: mtools module add-cli --module=example --name=hello-world --silent
`,
		Action: addCli.Invoke,
		Flags: []cli.Flag{
			flag.NewModule("A module name to add a CLI command to"),
			&cli.StringFlag{
				Name:    "name",
				Usage:   "The name of the CLI command",
				Aliases: []string{"n"},
			},
			flag.NewSilent("Do not ask for any input"),
		},
	}
}

func (a *AddCli) Invoke(ctx *cli.Context) error {
	mod, err := flag.ModuleValue(ctx)
	if err != nil {
		return nil
	}
	isSilent := flag.SilentValue(ctx)
	projPath := flag.ProjPathValue(ctx)

	name := ctx.String("name")
	if name == "" {
		if isSilent {
			fmt.Println(color.RedString("The command name is required"))
			return nil
		}
		name = a.askCommandName()
		if name == "" {
			return nil
		}
	}

	fmt.Println(
		color.GreenString("Adding a CLI command"),
		color.BlueString(name),
		color.GreenString(
			"to the module %s",
			color.BlueString(mod.Name),
		),
	)

	structName := strcase.ToCamel(name)

	err = utils.CreateDirIfNotExists(mod.CliPath(projPath))
	if err != nil {
		fmt.Println(
			color.RedString("Cannot create the CLI directory"),
			color.BlueString(mod.CliPath(projPath)),
			color.RedString(": %s", err.Error()),
		)
		return nil
	}

	err = a.createCommandFile(structName, name, mod, projPath)
	if err != nil {
		return nil
	}

	return nil
}

func (a *AddCli) createCommandFile(
	structName string,
	commandName string,
	mod module.ManifestModule,
	projPath string,
) error {
	path := mod.CliPath(projPath)
	pckg := mod.CliPackage()
	tmplVars := AddCliTmplVars{
		StructName:  structName,
		CommandName: commandName,
	}

	commandFile := strcase.ToSnake(commandName)

	err := utils.ProcessTemplate(
		"command.go.tmpl",
		"add_cli/command.go.tmpl",
		path+"/"+commandFile+".go",
		tmplVars,
	)
	if err != nil {
		fmt.Println(
			color.RedString("Cannot create the CLI command: %s", err.Error()),
		)
		return err
	}

	moduleFile := mod.ModulePath(projPath) + "/module.go"
	// this call is not necessary, but it's fine to have an import with defined alias
	_, err = files.AddImportToGoFile(pckg, "cmd", moduleFile)
	if err != nil {
		fmt.Println(
			color.RedString("Cannot add an import to the module.go file: %s", err.Error()),
		)
		return err
	}

	err = files.AddConstructorToProvider(pckg, "New"+structName, moduleFile)
	if err != nil {
		fmt.Println(
			color.RedString("Cannot add a constructor to the module.go file: %s", err.Error()),
		)
		return err
	}

	return nil
}

func (a *AddCli) askModuleName(
	manifest *module.Manifest,
) (string, error) {
	items := make([]string, 0)
	for _, md := range manifest.LocalModules() {
		items = append(items, md.Name)
	}
	sel := promptui.Select{
		Label: "Select a module",
		Items: items,
	}

	_, val, err := sel.Run()
	if err != nil {
		fmt.Println(color.RedString("Cannot ask module name: %s", err.Error()))
		return "", err
	}

	return val, nil
}

func (a *AddCli) askCommandName() string {
	for {
		prompt := promptui.Prompt{
			Label: "Enter a command name in the kebab-case. Example: hello-world",
		}

		name, err := prompt.Run()
		if err != nil {
			fmt.Println(color.RedString("Cannot ask command name: %s", err.Error()))
			return ""
		}
		if name == "" {
			fmt.Println(color.RedString("The command name cannot be empty"))
			continue
		}
		if !nameRegEx.MatchString(name) {
			fmt.Println(color.RedString("The command name must be in the kebab-case. Latin letters, numbers, and hyphens are allowed."))
			fmt.Println(color.BlueString("Example: hello-world"))
			continue
		}
		return name
	}
}
