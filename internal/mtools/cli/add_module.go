package cli

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errbuilder"
	"github.com/go-modulus/modulus/internal/mtools/files"
	"github.com/go-modulus/modulus/internal/mtools/utils"
	"github.com/go-modulus/modulus/module"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

var ErrPackageIsEmpty = errbuilder.New("package is empty").
	WithHint("Please provide a package for the module in the manifest file.").Build()
var ErrCannotRunGoGetCommand = errbuilder.New("cannot run go get command").Build()
var ErrCannotInstallModule = errbuilder.New("cannot install the module").
	WithHint("The install field in the manifest file should be a valid command running under 'go run'").Build()
var ErrCannotUpdateToolsFile = errbuilder.New("cannot update the tools file").
	WithHint("Check the existence and rights for the tools.go file at the root folder of your project.").Build()

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
Example: mtools add-module
Example without UI: mtools add-module --modules="urfave cli,pgx"
`,
		Action: addModule.Invoke,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "modules",
				Usage: "A comma-separated list of modules to add to the project",
			},
		},
	}
}

func (c *AddModule) Invoke(
	ctx *cli.Context,
) error {
	p := message.NewPrinter(language.English)
	projPath := ctx.String("proj-path")

	utils.PrintLogo()

	fmt.Println("Choose a module to add to your project")

	availableModulesManifest, err := module.NewFromFs(modulus.ManifestFs, "modules.json")
	if err != nil {
		fmt.Println(color.RedString("Cannot read from the manifest file: %s", err.Error()))
		return err
	}

	manifest, err := c.getLocalManifest()
	if err != nil {
		fmt.Println(color.RedString("Cannot get the local modules.json manifest file: %s", err.Error()))
		return err
	}

	fmt.Println(color.BlueString("Installed modules:"))
	for _, md := range manifest.Modules {
		fmt.Printf(
			"	%s: %s\n",
			color.BlueString(md.Name),
			md.Package,
		)
	}

	modulesValue := ctx.StringSlice("modules")
	var modules []module.ManifestItem
	if len(modulesValue) != 0 {
		for _, val := range modulesValue {
			for _, availableItem := range availableModulesManifest.Modules {
				if val == availableItem.Name {
					modules = append(modules, availableItem)
				}
			}
		}
	} else {
		modules, err = c.askModulesFromManifest(availableModulesManifest, manifest.Modules)
		if err != nil {
			fmt.Println(color.RedString("Cannot ask modules from the manifest: %s", err.Error()))
			return err
		}
	}
	if len(modules) == 0 {
		fmt.Println(color.YellowString("No modules were chosen. Exiting..."))
		return nil
	}

	entrypoints, err := c.getEntrypoints()
	if err != nil {
		fmt.Println(color.RedString("Cannot get the entrypoints: %s", err.Error()))
		return err
	}
	if len(entrypoints) == 0 {
		fmt.Println(
			color.YellowString(
				"No entrypoints were found. Please create a cmd folder with the entrypoints. \n" +
					"For example, you can create a cmd/console/main.go file with the main function. \n" +
					"Then run the command again. Exiting...",
			),
		)
		return nil
	}
	hasErrors := false
	for _, md := range modules {
		err = c.installModule(ctx.Context, md, entrypoints)
		if err != nil {
			fmt.Println(color.RedString("Cannot install the module %s: %s", md.Name, err.Error()))
			if errors.Hint(p, err) != "" {
				fmt.Println(color.YellowString("Hint: %s", errors.Hint(p, err)))
			}
			hasErrors = true
			continue
		}
		manifest.Modules = append(manifest.Modules, md)
		err = manifest.SaveAsLocalManifest(projPath)
		if err != nil {
			fmt.Println(color.RedString("Cannot save the local manifest file modules.json: %s", err.Error()))
			hasErrors = true
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

func (c *AddModule) saveLocalManifest(
	manifest module.Manifest,
) error {
	data, err := manifest.WriteToJSON()
	if err != nil {
		return err
	}
	return os.WriteFile("modules.json", data, 0644)
}

func (c *AddModule) getLocalManifest() (module.Manifest, error) {
	res := module.Manifest{
		Modules:     make([]module.ManifestItem, 0),
		Version:     "1.0.0",
		Name:        "Modulus framework modules manifest",
		Description: "List of installed modules for the Modulus framework",
	}
	if utils.FileExists("modules.json") {
		projFs := os.DirFS("./")
		manifest, err := module.NewFromFs(projFs, "modules.json")
		if err != nil {
			return res, err
		}
		return *manifest, nil
	}
	return res, nil
}

func (c *AddModule) installModule(
	ctx context.Context,
	md module.ManifestItem,
	entrypoints []entripoint,
) error {

	if md.Package == "" {
		return ErrPackageIsEmpty
	}

	fmt.Println(color.BlueString("Getting a package %s...", md.Package))
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	err := exec.CommandContext(cmdCtx, "go", "get", md.Package).Run()
	if err != nil {
		return errors.WrapCause(ErrCannotRunGoGetCommand, err)
	}

	fmt.Println(color.BlueString("Adding the package %s to the tools.go file...", md.Package))
	err = files.AddImportToTools(md.Package)
	if err != nil {
		return errors.WrapCause(ErrCannotUpdateToolsFile, err)
	}

	for _, entrypoint := range entrypoints {
		fmt.Println(color.BlueString("Adding module initialization to the entrypoint %s ...", entrypoint.name))
		err = files.AddModuleToEntrypoint(md.Package, entrypoint.path)
		if err != nil {
			fmt.Println(
				color.RedString(
					"Cannot add the module %s to the entrypoint %s: %s. Try to type initialization code manually",
					md.Name,
					entrypoint.path,
					err.Error(),
				),
			)
			continue
		}
		fmt.Println(color.BlueString("File %s is updated", entrypoint.path))
	}

	fmt.Println(color.BlueString("Running go mod tidy..."))
	err = exec.CommandContext(cmdCtx, "go", "mod", "tidy").Run()
	if err != nil {
		return errors.WrapCause(ErrCannotRunGoGetCommand, err)
	}

	if md.InstallCommand != "" {
		fmt.Println(
			color.BlueString(
				"Running the install command '%s' for the module %s...",
				md.InstallCommand,
				md.Name,
			),
		)
		cmdCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		err := exec.CommandContext(cmdCtx, "go", "run", md.InstallCommand).Run()
		if err != nil {
			return errors.WrapCause(ErrCannotInstallModule, err)
		}
	}

	fmt.Println(color.GreenString("The module %s has been successfully installed.", md.Name))
	return nil
}

func (c *AddModule) askModulesFromManifest(
	availableModulesManifest *module.Manifest,
	installedModules []module.ManifestItem,
) ([]module.ManifestItem, error) {
	res := make([]module.ManifestItem, 0)
	resNames := make([]string, 0)
	fmt.Println(color.BlueString(availableModulesManifest.Name))

	templates := &promptui.SelectTemplates{
		Label:    "{{if .IsSelected}}\U0001F4E6 {{ .Name | blue | faint }}{{else}}{{ . }}{{end}}",
		Active:   "→ {{if .IsSelected}}\U0001F4E6 {{end}} {{ .Name | cyan }} {{if .IsInstalled}}(installed){{end}}",
		Inactive: "{{if .IsSelected}}\U0001F4E6 {{end}} {{ .Name | white | faint }} {{if .IsInstalled}}(installed){{end}}",
		Details: `{{ if eq .Name  "Install chosen" }}{{ else }}
{{ "Package:" | faint }}	{{ .Package }}
{{ "Description:" | faint }}	{{ .Description }}{{ end }}`,
	}

	type selectItem struct {
		Name        string
		Package     string
		Description string
		IsSelected  bool
		IsInstalled bool
	}

	selectItems := make([]selectItem, len(availableModulesManifest.Modules)+1)
	selectItems[0] = selectItem{
		Name: "Install chosen",
	}
	for i, md := range availableModulesManifest.Modules {
		isInstalled := false
		for _, imd := range installedModules {
			if imd.Package == md.Package {
				isInstalled = true
				break
			}
		}
		selectItems[i+1] = selectItem{
			Name:        md.Name,
			Package:     md.Package,
			Description: md.Description,
			IsInstalled: isInstalled,
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

type entripoint struct {
	name string
	path string
}

func (c *AddModule) getEntrypoints() (entripoints []entripoint, err error) {
	entries, err := os.ReadDir("./cmd")
	if err != nil {
		return
	}
	entripoints = make([]entripoint, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			entryItem := entripoint{
				name: entry.Name(),
			}
			_, err2 := os.Stat("./cmd/" + entry.Name() + "/main.go")
			if os.IsNotExist(err2) {
				continue
			}

			if err2 != nil {
				err = err2
				fmt.Println(color.RedString("Error when getting an entrypoint %s: %s", entry.Name(), err.Error()))
				return
			}
			entryItem.path = "./cmd/" + entry.Name() + "/main.go"
			entripoints = append(entripoints, entryItem)
		}
	}

	return
}
