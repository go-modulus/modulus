package cli

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/internal/mtools/action"
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
	"slices"
	"strings"
	"time"
)

var moduleNameRegexp = regexp.MustCompile(`module\s+([a-zA-Z0-9_\-\/]+)+`)

type features struct {
	storage bool
	graphQL bool
}

type CreateModule struct {
	logger         *slog.Logger
	installStorage *action.InstallStorage
}

func NewCreateModule(
	logger *slog.Logger,
	installStorage *action.InstallStorage,
) *CreateModule {
	return &CreateModule{
		logger:         logger,
		installStorage: installStorage,
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
			&cli.BoolFlag{
				Name:  "silent",
				Usage: "Set the silent mode to disable asking the questions",
			},
			&cli.StringSliceFlag{
				Name:  "without",
				Usage: "Set the list of features to install the module without. Available values: storage, graphql",
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

	projPath := ctx.String("proj-path")
	err = c.saveManifestItem(manifestItem, projPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(manifestItem.LocalPath, 0755)
	if err != nil {
		fmt.Println(color.RedString("Cannot create a directory %s: %s", manifestItem.LocalPath, err.Error()))
		return err
	}

	selectedFeatures := c.getFeatures(ctx)

	if selectedFeatures.storage {
		err = c.installStorageFeature(ctx, manifestItem, projPath)
		if err != nil {
			return err
		}
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

func (c *CreateModule) installStorageFeature(
	ctx *cli.Context,
	md module.ManifestItem,
	projPath string,
) error {
	cfg := action.StorageConfig{
		Schema:             "public",
		GenerateGraphql:    true,
		GenerateFixture:    true,
		GenerateDataloader: true,
		ProjPath:           projPath,
	}
	if !ctx.Bool("silent") {
		schema, err := c.askSchema(cfg.Schema)
		if err != nil {
			return err
		}
		cfg.Schema = schema

		cfg.GenerateGraphql, err = c.askYesNo("Do you want to generate GraphQL files from SQL?")
		if err != nil {
			return err
		}
		cfg.GenerateFixture, err = c.askYesNo("Do you want to generate fixture files from SQL?")
		if err != nil {
			return err
		}
		cfg.GenerateDataloader, err = c.askYesNo("Do you want to generate dataloader files from SQL?")
		if err != nil {
			return err
		}
	}
	return c.installStorage.Install(ctx.Context, md, cfg)
}

func (c *CreateModule) getFeatures(ctx *cli.Context) (res features) {
	res = features{
		storage: true,
		graphQL: true,
	}
	without := ctx.StringSlice("without")
	type feature struct {
		name        string
		description string
		value       *bool
	}
	items := []feature{
		{
			name: "storage",
			description: "The storage feature allows you to work with PostgreSQL.\n" +
				"It includes migrations and SQLc generated files to call DB queries.",
			value: &res.storage,
		},
		{
			name: "graphql",
			description: "The GraphQL feature allows you to work with GraphQL.\n" +
				"It includes the resolvers and GraphQL schemas compatible with gqlgen.",
			value: &res.graphQL,
		},
	}
	for _, w := range without {
		switch w {
		case "storage":
			res.storage = false
			items = slices.DeleteFunc(
				items, func(val feature) bool {
					return val.name == "storage"
				},
			)
		case "graphql":
			res.graphQL = false
			items = slices.DeleteFunc(
				items, func(val feature) bool {
					return val.name == "graphql"
				},
			)
		}
	}
	if len(items) != 0 && !ctx.Bool("silent") {
		for _, item := range items {
			val, err := c.askYesNo("Do you want to install the " + item.name + " feature?\n" + item.description)
			if err != nil {
				return
			}
			*item.value = val
		}

	}
	return
}

func (c *CreateModule) askSchema(defSchema string) (string, error) {
	prompt := promptui.Prompt{
		Label:   "Enter a PG schema where you want to place tables for this module: ",
		Default: defSchema,
	}

	return prompt.Run()
}

func (c *CreateModule) askYesNo(label string) (bool, error) {
	sel := promptui.Select{
		Label: label,
		Items: []string{"Yes", "No"},
	}
	_, result, err := sel.Run()
	if err != nil {
		fmt.Println(color.RedString("Cannot ask a question: %s", err.Error()))
		return false, err
	}
	val := false
	if result == "Yes" {
		val = true
	}
	return val, nil
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

func (c *CreateModule) saveManifestItem(manifestItem module.ManifestItem, projPath string) (err error) {
	manifest, err := module.LoadLocalManifest(projPath)
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
	err = manifest.SaveAsLocalManifest(projPath)
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
	isSilent := ctx.Bool("silent")
	pckg := ctx.String("package")
	if pckg == "" {
		if isSilent {
			fmt.Println(color.RedString("The package name is not provided. Please add the --package flag or remove the --silent=true flag"))
			return module.ManifestItem{}, errors.New("the package name is not provided")
		}
		pckg, err = c.askPackage()
	}

	name := ctx.String("name")
	if name == "" {
		if !isSilent {
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
		if !isSilent {
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
		IsLocalModule:  true,
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
