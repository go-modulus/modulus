package cli

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/internal/mtools/files"
	"github.com/go-modulus/modulus/internal/mtools/templates"
	"github.com/go-modulus/modulus/internal/mtools/utils"
	"github.com/go-modulus/modulus/module"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"html/template"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var moduleNameRegexp = regexp.MustCompile(`module\s+([a-zA-Z0-9_\-\/]+)+`)

type CreateModule struct {
	logger *slog.Logger
}

func NewCreateModule(
	logger *slog.Logger,
) *CreateModule {
	return &CreateModule{
		logger: logger,
	}
}

func NewCreateModuleCommand(createModule *CreateModule) *cli.Command {
	return &cli.Command{
		Name: "create-module",
		Usage: `Create a boilerplate of the new module and place its files inside the obtained path.
Adds the chosen module to the project and inits it with copying necessary files.
Example: mtools create-module
Example without UI: mtools create-module --path=internal/mypckg --package=mypckg --name="My package"
Example filling default values without UI: mtools create-module --package=mypckg
`,
		Action: createModule.Invoke,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "package",
				Usage: "A package name of the module Go file",
			},
			&cli.StringFlag{
				Name:  "path",
				Usage: "A local path to the module",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "A name of the module",
			},
		},
	}
}

func (c *CreateModule) Invoke(
	ctx *cli.Context,
) (err error) {
	utils.PrintLogo()

	manifestItem, err := c.getManifestItem(ctx)
	if err != nil {
		return err
	}

	err = c.saveManifestItem(manifestItem)
	if err != nil {
		return err
	}

	err = os.MkdirAll(manifestItem.LocalPath, 0755)
	if err != nil {
		fmt.Println(color.RedString("Cannot create a directory %s: %s", manifestItem.LocalPath, err.Error()))
		return err
	}

	err = c.addModuleFile(manifestItem)
	if err != nil {
		return err
	}
	fmt.Println(
		"Congratulations! Your module is created.",
	)

	return nil
}

func (c *CreateModule) addModuleFile(
	md module.ManifestItem,
) error {
	tmpl := template.Must(
		template.New("module.go.tmpl").
			ParseFS(
				templates.TemplateFiles,
				"create_module/module.go.tmpl",
			),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := tmpl.ExecuteTemplate(w, "module.go.tmpl", &md)
	if err != nil {
		return err
	}
	w.Flush()

	err = os.WriteFile(md.LocalPath+"/module.go", b.Bytes(), 0644)
	if err != nil {
		fmt.Println(color.RedString("Cannot write a module file: %s", err.Error()))
		return err
	}
	return nil
}

func (c *CreateModule) saveManifestItem(manifestItem module.ManifestItem) (err error) {
	manifest, err := module.LoadLocalManifest()
	if err != nil {
		fmt.Println(color.RedString("Cannot get a local manifest: %s", err.Error()))
		return err
	}
	for _, item := range manifest.Modules {
		if item.Package == manifestItem.Package {
			fmt.Println(color.YellowString("The module %s is already installed", item.Name))
			return err
		}
	}
	manifest.Modules = append(
		manifest.Modules, manifestItem,
	)
	err = manifest.SaveAsLocalManifest()
	if err != nil {
		fmt.Println(color.RedString("Cannot save a local manifest: %s", err.Error()))
		return err
	}
	return nil
}

func (c *CreateModule) getProjModuleName() (string, error) {
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

func (c *CreateModule) getManifestItem(ctx *cli.Context) (
	res module.ManifestItem,
	err error,
) {
	pckg := ctx.String("package")
	obtainedPckg := true
	if pckg == "" {
		obtainedPckg = false
		pckg, err = c.askPackage()
	}

	name := ctx.String("name")
	if name == "" {
		if !obtainedPckg {
			name, err = c.askName(pckg)
			if err != nil {
				fmt.Println(color.RedString("Cannot ask a name: %s", err.Error()))
				return module.ManifestItem{}, err
			}
		} else {
			// If we are in a silence mode, we need to get a name from the package
			name = pckg
		}
	}

	path := ctx.String("path")
	if path == "" {
		if !obtainedPckg {
			path, err = c.askPath(pckg)
			if err != nil {
				fmt.Println(color.RedString("Cannot ask a path: %s", err.Error()))
				return module.ManifestItem{}, err
			}
		} else {
			path = c.getDefaultPath(pckg)
		}
	}

	projPckg, err := c.getProjModuleName()
	if err != nil {
		return module.ManifestItem{}, err
	}

	res = module.ManifestItem{
		Name:           name,
		Package:        projPckg + "/" + path,
		Description:    "",
		InstallCommand: "",
		Version:        "",
		LocalPath:      path,
	}
	return res, nil
}

func (c *CreateModule) saveLocalManifest(
	manifest module.Manifest,
) error {
	data, err := manifest.WriteToJSON()
	if err != nil {
		return err
	}
	return os.WriteFile("modules.json", data, 0644)
}

func (c *CreateModule) installModule(
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

func (c *CreateModule) askPath(packageName string) (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter a folder starting from the root of a project: ",
	}

	suggestion := c.getDefaultPath(packageName)
	prompt.Default = suggestion

	path, err := prompt.Run()
	if err != nil {
		return "", err
	}
	path = "./" + path

	return path, nil
}

func (c *CreateModule) askName(packageName string) (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter a name of the module: ",
	}

	prompt.Default = packageName

	return prompt.Run()
}

func (c *CreateModule) getDefaultPath(packageName string) string {
	nameParts := strings.Split(packageName, "/")
	return "internal/" + nameParts[len(nameParts)-1]
}

func (c *CreateModule) askPackage() (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter a Golang package name of the created module (e.g. user): ",
	}

	pckg, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return pckg, nil
}

func (c *CreateModule) getEntrypoints() (entripoints []entripoint, err error) {
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
