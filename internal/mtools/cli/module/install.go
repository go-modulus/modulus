package module

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errbuilder"
	"github.com/go-modulus/modulus/internal/mtools/cli/flag"
	"github.com/go-modulus/modulus/internal/mtools/files"
	"github.com/go-modulus/modulus/internal/mtools/utils"
	"github.com/go-modulus/modulus/module"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"
)

var ErrPackageIsEmpty = errbuilder.New("package is empty").
	WithHint("Please provide a package for the module in the manifest file.").Build()
var ErrCannotRunGoGetCommand = errbuilder.New("cannot run go get command").Build()
var ErrCannotInstallModule = errbuilder.New("cannot install the module").
	WithHint("The install field in the manifest file should be a valid command running under 'go run'").Build()
var ErrCannotUpdateToolsFile = errbuilder.New("cannot update the tools file").
	WithHint("Check the existence and rights for the tools.go file at the root folder of your project.").Build()

type InstalledFileVars struct {
	ModuleName     string
	ProjectPackage string
}

type Install struct {
	logger *slog.Logger
}

func NewInstall(
	logger *slog.Logger,
) *Install {
	return &Install{
		logger: logger,
	}
}

func NewInstallCommand(addModule *Install) *cli.Command {
	return &cli.Command{
		Name: "install",
		Usage: `Gives user a choice to install any modules from the available ones list.
Uses interactive prompts to make a choice.
Adds the chosen module to the project and inits it with default files.
Example: mtools module install
Example without UI: mtools module install --modules="urfave cli,pgx"
Example with a custom manifest located at https://example.com/modules.json: mtools module install --manifest="https://example.com"
`,
		Action: addModule.Invoke,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "modules",
				Usage:   "A comma-separated list of modules names to add to the project",
				Aliases: []string{"m"},
			},
			flag.NewManifest(
				`A path to the global manifest with all available modules to install.
Example: mtools module install --manifest="local_folder/modules.json"`,
			),
		},
	}
}

func (c *Install) Invoke(
	ctx *cli.Context,
) error {
	p := message.NewPrinter(language.English)
	modulesValue := ctx.StringSlice("modules")
	if len(modulesValue) == 0 {
		utils.PrintLogo()
	}
	availableModulesManifest, err := flag.ManifestValue(ctx)
	if err != nil {
		fmt.Println(color.RedString("Cannot get the manifest file: %s", err.Error()))
		return err
	}
	projPath := ctx.String("proj-path")
	if projPath != "" {
		curDir, err := os.Getwd()
		if err != nil {
			fmt.Println(color.RedString("Cannot get the current directory: %s", err.Error()))
			return err
		}
		err = os.Chdir(projPath)
		if err != nil {
			fmt.Println(color.RedString("Cannot change the current directory to %s: %s", projPath, err.Error()))
			return err
		}
		curDirAfterChange, _ := os.Getwd()
		fmt.Printf("Changing the current dir to %s\n", color.BlueString(curDirAfterChange))
		defer func() { _ = os.Chdir(curDir) }()
	}

	manifest, err := c.getLocalManifest()
	if err != nil {
		fmt.Println(color.RedString("Cannot get the local modules.json manifest file: %s", err.Error()))
		return err
	}

	fmt.Println("Installed modules:")
	for _, md := range manifest.Modules {
		fmt.Printf(
			"	%s: %s\n",
			color.BlueString(md.Name),
			md.Package,
		)
	}

	var modules []module.ManifestModule
	if len(modulesValue) != 0 {
		fmt.Printf("Modules to install: %s\n", color.BlueString(strings.Join(modulesValue, ", ")))
		for _, val := range modulesValue {
			for _, availableItem := range availableModulesManifest.Modules {
				if val == availableItem.Name {
					modules = append(modules, availableItem)
				}
			}
		}
	} else {
		fmt.Println("Choose a module to add to your project")
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

	modules, err = c.addDependedModulesToInstall(availableModulesManifest, manifest.Modules, modules)
	if err != nil {
		fmt.Println(color.RedString("Cannot add depended modules to the install: %s", err.Error()))
		return err
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
	localModulesMap := make(map[string]struct{})
	for _, md := range manifest.Modules {
		localModulesMap[md.Package] = struct{}{}
	}
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
		if _, ok := localModulesMap[md.Package]; !ok {
			manifest.Modules = append(manifest.Modules, md)

			err = manifest.SaveAsLocalManifest("./")
			if err != nil {
				fmt.Println(color.RedString("Cannot save the local manifest file modules.json: %s", err.Error()))
				hasErrors = true
			}
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

func (c *Install) getLocalManifest() (module.Manifest, error) {
	res := module.Manifest{
		Modules:     make([]module.ManifestModule, 0),
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

func (c *Install) installModule(
	ctx context.Context,
	md module.ManifestModule,
	entrypoints []entripoint,
) error {

	if md.Package == "" {
		return ErrPackageIsEmpty
	}

	fmt.Printf("Getting a package %s...\n", color.BlueString(md.Package))
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	err := exec.CommandContext(cmdCtx, "go", "get", md.Package).Run()
	if err != nil {
		return errors.WrapCause(ErrCannotRunGoGetCommand, err)
	}

	fmt.Printf("Adding the package %s to the tools.go file...\n", color.BlueString(md.Package))
	err = files.AddImportToTools(md.Package)
	if err != nil {
		return errors.WrapCause(ErrCannotUpdateToolsFile, err)
	}

	for _, entrypoint := range entrypoints {
		fmt.Printf("Adding module initialization to the entrypoint %s ...\n", color.BlueString(entrypoint.name))
		err = files.AddModuleToEntrypoint(md.Package, entrypoint.path)
		if err != nil {
			fmt.Println(
				color.RedString(
					"Cannot add the module %s to the entrypoint %s: %s. Try to type initialization code manually",
					color.BlueString(md.Name),
					color.BlueString(entrypoint.path),
					err.Error(),
				),
			)
			continue
		}
		fmt.Printf("File %s is updated\n", color.BlueString(entrypoint.path))
	}

	fmt.Printf("Running %s...\n", color.BlueString("go mod tidy"))
	err = exec.CommandContext(cmdCtx, "go", "mod", "tidy").Run()
	if err != nil {
		return errors.WrapCause(ErrCannotRunGoGetCommand, err)
	}

	if len(md.Install.EnvVars) != 0 {
		err = module.WriteEnvVariablesToFile(md.Install.EnvVars, ".env")
		if err != nil {
			fmt.Println("Cannot update the .env file:", color.RedString(err.Error()))
			return err
		}
	}
	if len(md.Install.Files) != 0 {
		fmt.Println("Downloading the files...")
		for _, file := range md.Install.Files {
			fmt.Printf("Downloading the file %s...\n", color.BlueString(file.SourceUrl))
			err = c.copyRemoteFile(md, file)
			if err != nil {
				fmt.Println("Cannot download the file:", color.RedString(err.Error()))
				return err
			}
		}
	}
	if md.LocalPath != "" {
		fmt.Println("Initializing the local module...")
		projPackage, err := c.getProjPackage()
		if err != nil {
			return err
		}
		localModulePackage := projPackage + "/" + md.LocalPath
		for _, entrypoint := range entrypoints {
			fmt.Printf("Adding module initialization to the entrypoint %s ...\n", color.BlueString(entrypoint.name))
			err = files.AddModuleToEntrypoint(localModulePackage, entrypoint.path)
			if err != nil {
				fmt.Println(
					color.RedString(
						"Cannot add the module %s to the entrypoint %s: %s. Try to type initialization code manually",
						color.BlueString(md.Name),
						color.BlueString(entrypoint.path),
						err.Error(),
					),
				)
				continue
			}
			fmt.Printf("File %s is updated\n", color.BlueString(entrypoint.path))
		}
	}
	if len(md.Install.PostInstallCommands) != 0 {
		fmt.Printf("Running the post install commands for the module %s...\n", color.BlueString(md.Name))
		for _, cmd := range md.Install.PostInstallCommands {
			fmt.Printf("Adding the package %s to the tools.go file...\n", color.BlueString(md.Package))
			err = exec.CommandContext(cmdCtx, "go", "get", cmd.CmdPackage).Run()
			if err != nil {
				return errors.WrapCause(ErrCannotRunGoGetCommand, err)
			}
			err = files.AddImportToTools(cmd.CmdPackage)
			if err != nil {
				return errors.WrapCause(ErrCannotUpdateToolsFile, err)
			}
			fmt.Printf("Running %s...\n", color.BlueString("go run "+cmd.CmdPackage))
			params := append([]string{"run", cmd.CmdPackage}, cmd.Params...)
			err = exec.CommandContext(cmdCtx, "go", params...).Run()
			if err != nil {
				return errors.WrapCause(ErrCannotInstallModule, err)
			}
		}
	}

	fmt.Println(color.GreenString("The module %s has been successfully installed.", color.BlueString(md.Name)))
	return nil
}

func (c *Install) copyRemoteFile(
	md module.ManifestModule,
	file module.InstalledFile,
) error {
	if utils.FileExists(file.DestFile) {
		fmt.Println(
			color.YellowString("The file"),
			color.BlueString(file.DestFile),
			color.YellowString("already exists. Skipping..."),
		)
		return nil
	}
	//download file
	resp, err := http.Get(file.SourceUrl)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return err
	}

	var rb bytes.Buffer
	_, err = rb.ReadFrom(resp.Body)
	if err != nil {
		fmt.Println("Cannot read the response body:", color.RedString(err.Error()))
		return err
	}
	tpl := `{{define "main"}}` + rb.String() + `{{end}}`

	tmpl := template.Must(
		template.New("main").
			Parse(
				tpl,
			),
	)

	projPackage, err := c.getProjPackage()
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	vars := InstalledFileVars{
		ModuleName:     md.Name,
		ProjectPackage: projPackage,
	}
	err = tmpl.ExecuteTemplate(w, "main", &vars)
	if err != nil {
		return err
	}
	w.Flush()

	dir := path.Dir(file.DestFile)
	if dir != "." && dir != "" && dir != "/" {
		fmt.Println("Preparing directory for the file...")
		err = utils.CreateDirIfNotExists(dir)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(file.DestFile, b.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *Install) getProjPackage() (string, error) {
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println(color.RedString("The go.mod file is not found. Try to run the command in the root of the project"))
		return "", err
	}
	content, err := os.ReadFile("go.mod")
	if err != nil {
		fmt.Println(color.RedString("Cannot read a go.mod file: %s", err.Error()))
		return "", err
	}

	moduleStr := moduleNameRegexp.FindStringSubmatch(string(content))
	if len(moduleStr) < 2 {
		fmt.Println(color.RedString("Cannot find a module name in the go.mod file"))
		return "", errors.New("cannot find a module name in the go.mod file")
	}

	return moduleStr[1], nil
}

func (c *Install) addDependedModulesToInstall(
	availableModulesManifest *module.Manifest,
	installedModules []module.ManifestModule,
	modulesToInstall []module.ManifestModule,
) ([]module.ManifestModule, error) {
	addedModules := make([]module.ManifestModule, 0)
	addedModulesMap := make(map[string]struct{})
	availableModulesMap := make(map[string]module.ManifestModule)
	for _, md := range availableModulesManifest.Modules {
		availableModulesMap[md.Name] = md
	}
	for _, md := range installedModules {
		addedModulesMap[md.Name] = struct{}{}
	}
	for _, md := range modulesToInstall {
		addedModulesMap[md.Package] = struct{}{}
	}

	for _, md := range modulesToInstall {
		if len(md.Install.Dependencies) == 0 {
			continue
		}
		for _, depName := range md.Install.Dependencies {
			if availableDep, ok := availableModulesMap[depName]; ok {
				if _, ok := addedModulesMap[availableDep.Name]; !ok {
					addedModules = append(addedModules, availableDep)
					addedModulesMap[availableDep.Name] = struct{}{}
				}
			}
		}
	}

	return append(modulesToInstall, addedModules...), nil
}

func (c *Install) askModulesFromManifest(
	availableModulesManifest *module.Manifest,
	installedModules []module.ManifestModule,
) ([]module.ManifestModule, error) {
	res := make([]module.ManifestModule, 0)
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
			Size:      15,
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

func (c *Install) getEntrypoints() (entripoints []entripoint, err error) {
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
