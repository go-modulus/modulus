package cli

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errbuilder"
	"github.com/go-modulus/modulus/internal/mtools/utils"
	"github.com/go-modulus/modulus/module"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

var ErrPackageIsEmpty = errbuilder.New("package is empty").
	WithHint("Please provide a package for the module in the manifest file.").Build()
var ErrCannotRunGoGetCommand = errbuilder.New("cannot run go get command").Build()
var ErrCannotInstallModule = errbuilder.New("cannot install the module").
	WithHint("The install field in the manifest file should be a valid command running under 'go run'").Build()

type AddModule struct {
	logger *slog.Logger
}

func NewAddModule(
	logger *slog.Logger,
) *AddModule {
	return &AddModule{
		logger: logger,
	}
}

func NewAddModuleCommand(addModule *AddModule) *cli.Command {
	return &cli.Command{
		Name: "add-module",
		Usage: `Gives user a choice to get any modules from the available ones list.
Uses interactive prompts to make a choice.
Adds the chosen module to the project and inits it with copying necessary files.
Example: ./bin/modulus add-module
`,
		Action: addModule.Invoke,
	}
}

func (c *AddModule) Invoke(
	ctx *cli.Context,
) error {
	p := message.NewPrinter(language.English)

	utils.PrintLogo()

	fmt.Println("Choose a module to add to your project")

	availableModulesManifest, err := module.NewFromFs(modulus.ManifestFs, "modules.json")
	if err != nil {
		fmt.Println(color.RedString("Cannot read from the manifest file: %s", err.Error()))
		return err
	}

	modules, err := c.askModulesFromManifest(availableModulesManifest)
	if err != nil {
		fmt.Println(color.RedString("Cannot ask modules from the manifest: %s", err.Error()))
		return err
	}
	if len(modules) == 0 {
		fmt.Println(color.YellowString("No modules were chosen. Exiting..."))
		return nil
	}

	hasErrors := false
	for _, md := range modules {
		err = c.installModule(ctx.Context, md)
		if err != nil {
			fmt.Println(color.RedString("Cannot install the module %s: %s", md.Name, err.Error()))
			if errors.Hint(p, err) != "" {
				fmt.Println(color.YellowString("Hint: %s", errors.Hint(p, err)))
			}
			hasErrors = true
			continue
		}
	}
	if hasErrors {
		fmt.Println(color.YellowString("Some modules were not installed. Exiting..."))
		return nil
	}
	fmt.Println(
		"Congratulations! Your project has been updated.",
	)

	return nil
}

func (c *AddModule) installModule(
	ctx context.Context,
	md module.ManifestItem,
) error {

	if md.Package == "" {
		return ErrPackageIsEmpty
	}
	if md.Package != "github.com/go-modulus/modulus" {
		cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
		err := exec.CommandContext(cmdCtx, "go", "get", md.Package).Run()
		if err != nil {
			return errors.WrapCause(ErrCannotRunGoGetCommand, err)
		}
	}
	if md.InstallCommand != "" {
		cmdCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		err := exec.CommandContext(cmdCtx, "go", "run", md.InstallCommand).Run()
		if err != nil {
			return errors.WrapCause(ErrCannotInstallModule, err)
		}
	}

	return nil
}

func (c *AddModule) askModulesFromManifest(
	availableModulesManifest *module.Manifest,
) ([]module.ManifestItem, error) {
	res := make([]module.ManifestItem, 0)
	resNames := make([]string, 0)
	fmt.Println(color.BlueString(availableModulesManifest.Name))

	templates := &promptui.SelectTemplates{
		Label:    "{{if .IsSelected}}\U0001F4E6 {{ .Name | blue | faint }}{{else}}{{ . }}{{end}}",
		Active:   "→ {{if .IsSelected}}\U0001F4E6 {{end}} {{ .Name | cyan }}",
		Inactive: "{{if .IsSelected}}\U0001F4E6 {{end}} {{ .Name | white | faint }}",
		//Selected: "\U0001F4E6 {{ .Name | blue | faint }}",
		Details: `{{ if eq .Name  "Exit" }}{{ else }}
{{ "Package:" | faint }}	{{ .Package }}
{{ "Description:" | faint }}	{{ .Description }}{{ end }}`,
	}

	type selectItem struct {
		Name        string
		Package     string
		Description string
		IsSelected  bool
	}

	selectItems := make([]selectItem, len(availableModulesManifest.Modules)+1)
	selectItems[0] = selectItem{
		Name: "Exit",
	}
	for i, md := range availableModulesManifest.Modules {
		selectItems[i+1] = selectItem{
			Name:        md.Name,
			Package:     md.Package,
			Description: md.Description,
		}
	}

	searcher := func(input string, index int) bool {
		curModule := availableModulesManifest.Modules[index-1]
		name := strings.Replace(strings.ToLower(curModule.Name), " ", "", -1)
		descr := strings.Replace(strings.ToLower(curModule.Description), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input) || strings.Contains(descr, input)
	}

	selectedPos := 0
	for {
		sel := promptui.Select{
			Label:     "Please choose a module from the list below:",
			Items:     selectItems,
			Templates: templates,
			Size:      5,
			Searcher:  searcher,
			// Start the cursor at the currently selected index
			CursorPos:    selectedPos,
			HideSelected: true,
		}

		index, _, err := sel.Run()
		if err != nil {
			return nil, err
		}
		if index == 0 {
			break
		}

		selectItems[index].IsSelected = !selectItems[index].IsSelected
		if selectItems[index].IsSelected {
			res = append(res, availableModulesManifest.Modules[index-1])
			resNames = append(resNames, availableModulesManifest.Modules[index-1].Name)
			selectedPos = index
		} else {
			for i, r := range res {
				if r.Name+r.Package == selectItems[index].Name+selectItems[index].Package {
					res = append(res[:i], res[i+1:]...)
					resNames = append(resNames[:i], resNames[i+1:]...)
					break
				}
			}
		}

	}
	if len(res) > 0 {
		fmt.Printf(
			"You have chosen the following modules:\n%s\n",
			color.BlueString(strings.Join(resNames, "\n")),
		)
	}
	return res, nil
}
