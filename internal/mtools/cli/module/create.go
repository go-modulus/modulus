package module

import (
	"bufio"
	"bytes"
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
	"regexp"
	"slices"
)

var moduleNameRegexp = regexp.MustCompile(`module\s+([a-zA-Z0-9_\-\/]+)+`)
var pckgNameRegexp = regexp.MustCompile(`^[a-z]+[a-z0-9]+`)

type features struct {
	storage bool
	graphQL bool
}

type TmplVars struct {
	Module     module.ManifestModule
	HasStorage bool
}

type Create struct {
	logger         *slog.Logger
	installStorage *action.InstallStorage
}

func NewCreate(
	logger *slog.Logger,
	installStorage *action.InstallStorage,
) *Create {
	return &Create{
		logger:         logger,
		installStorage: installStorage,
	}
}

func NewCreateCommand(createModule *Create) *cli.Command {
	return &cli.Command{
		Name: "create",
		Usage: `Create a boilerplate of the new module and place its files inside the obtained path.
Adds the chosen module to the project and inits it with copying necessary files.
Example: mtools module create
Example without UI: mtools module create --path=internal/mypckg --package=mypckg --name="My package"
Example filling default values without UI: mtools module create --package=mypckg
`,
		Action: createModule.Invoke,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "package",
				Usage:   "A package name of the module Go file",
				Aliases: []string{"pkg"},
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
				Name:    "silent",
				Usage:   "Set the silent mode to disable asking the questions",
				Aliases: []string{"s"},
			},
			&cli.StringSliceFlag{
				Name:  "without",
				Usage: "Set the list of features to install the module without. Available values: storage, graphql",
			},
		},
	}
}

func (c *Create) Invoke(
	ctx *cli.Context,
) (err error) {
	if !ctx.Bool("silent") {
		utils.PrintLogo()
	}
	projPath := ctx.String("proj-path")

	manifestItem, err := c.getManifestItem(ctx, projPath)
	if err != nil {
		return err
	}

	err = c.saveManifestItem(manifestItem, projPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(projPath+"/"+manifestItem.LocalPath, 0755)
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

	err = c.addModuleFile(manifestItem, projPath, selectedFeatures)
	if err != nil {
		return err
	}

	err = c.updateEntripoints(projPath, manifestItem)
	if err != nil {
		return err
	}

	fmt.Println(
		"Congratulations! Your module is created.",
	)

	return nil
}

func (c *Create) updateEntripoints(
	projPath string,
	md module.ManifestModule,
) error {
	fmt.Println(color.BlueString("Updating entrypoints..."))

	manifest, err := module.LoadLocalManifest(projPath)
	if err != nil {
		fmt.Println(color.RedString("Cannot get a local manifest: %s", err.Error()))
		return err
	}
	for _, entry := range manifest.Entries {
		err = files.AddModuleToEntrypoint(md.Package, projPath+"/"+entry.LocalPath)
		if err != nil {
			fmt.Println(
				color.RedString(
					"Cannot add the module %s to the entrypoint %s: %s. Try to type initialization code manually",
					md.Name,
					entry.LocalPath,
					err.Error(),
				),
			)
			continue
		}
	}

	return nil
}

func (c *Create) installStorageFeature(
	ctx *cli.Context,
	md module.ManifestModule,
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

func (c *Create) getFeatures(ctx *cli.Context) (res features) {
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

func (c *Create) askSchema(defSchema string) (string, error) {
	prompt := promptui.Prompt{
		Label:   "Enter a PG schema where you want to place tables for this module: ",
		Default: defSchema,
	}

	return prompt.Run()
}

func (c *Create) askYesNo(label string) (bool, error) {
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

func (c *Create) addModuleFile(
	md module.ManifestModule,
	projPath string,
	selectedFeatures features,
) error {
	vars := TmplVars{
		Module:     md,
		HasStorage: selectedFeatures.storage,
	}
	tmpl := template.Must(
		template.New("module.go.tmpl").
			ParseFS(
				templates.TemplateFiles,
				"create_module/module.go.tmpl",
			),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := tmpl.ExecuteTemplate(w, "module.go.tmpl", &vars)
	if err != nil {
		return err
	}
	w.Flush()

	err = os.WriteFile(md.ModulePath(projPath)+"/module.go", b.Bytes(), 0644)
	if err != nil {
		fmt.Println(color.RedString("Cannot write a module file: %s", err.Error()))
		return err
	}
	return nil
}

func (c *Create) saveManifestItem(manifestItem module.ManifestModule, projPath string) (err error) {
	manifest, err := module.LoadLocalManifest(projPath)
	if err != nil {
		fmt.Println(color.RedString("Cannot get a local manifest: %s", err.Error()))
		return err
	}
	for _, item := range manifest.Modules {
		if item.Package == manifestItem.Package {
			fmt.Println(color.YellowString("The module %s is already installed", item.Name))
			return errors.New("the module is already installed")
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

func (c *Create) getProjModuleName(projPath string) (string, error) {
	if _, err := os.Stat(projPath + "/go.mod"); os.IsNotExist(err) {
		fmt.Println(color.RedString("The go.mod file is not found. Try to run the command in the root of the project"))
		return "", err
	}
	content, err := os.ReadFile(projPath + "/go.mod")
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

func (c *Create) getManifestItem(ctx *cli.Context, projPath string) (
	res module.ManifestModule,
	err error,
) {
	isSilent := ctx.Bool("silent")
	pckg := ctx.String("package")
	if pckg == "" {
		if isSilent {
			fmt.Println(color.RedString("The package name is not provided. Please add the --package flag or remove the --silent=true flag"))
			return module.ManifestModule{}, errors.New("the package name is not provided")
		}
		pckg, err = c.askPackage()
	} else {
		if !pckgNameRegexp.MatchString(pckg) {
			fmt.Println(
				color.RedString(
					"The package name %s is not valid. Please use lowercase latin symbols without spaces",
					pckg,
				),
			)
			return module.ManifestModule{}, errors.New("the package name is not valid")
		}
	}

	name := ctx.String("name")
	if name == "" {
		if !isSilent {
			name, err = c.askName(pckg)
			if err != nil {
				fmt.Println(color.RedString("Cannot ask a name: %s", err.Error()))
				return module.ManifestModule{}, err
			}
		} else {
			// If we are in a silence mode, we need to get a name from the package
			name = pckg
		}
	}

	path := ctx.String("path")
	if path == "" {
		if !isSilent {
			path, err = c.askPath(projPath, pckg)
			if err != nil {
				fmt.Println(color.RedString("Cannot ask a path: %s", err.Error()))
				return module.ManifestModule{}, err
			}
		} else {
			path = c.getDefaultPath(pckg)
		}
	} else {
		path += "/" + pckg
	}

	projPckg, err := c.getProjModuleName(projPath)
	if err != nil {
		return module.ManifestModule{}, err
	}

	res = module.ManifestModule{
		Name:          name,
		Package:       projPckg + "/" + path,
		Description:   "",
		Version:       "",
		LocalPath:     path,
		IsLocalModule: true,
	}
	return res, nil
}

func (c *Create) askPath(projPath string, packageName string) (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter a folder starting from the root of a project (" +
			color.BlueString(projPath) + "): ",
	}

	suggestion := "internal"
	prompt.Default = suggestion

	path, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", path, packageName), nil
}

func (c *Create) askName(packageName string) (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter a name of the module: ",
	}

	prompt.Default = packageName

	return prompt.Run()
}

func (c *Create) getDefaultPath(packageName string) string {
	return fmt.Sprintf("internal/%s", packageName)
}

func (c *Create) askPackage() (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter a Golang package name of the created module (e.g. user): ",
	}

	var pckg string
	var err error
	for {
		pckg, err = prompt.Run()
		if err != nil {
			return "", err
		}
		if pckg == "" {
			fmt.Println(color.RedString("The package name cannot be empty"))
			continue
		}
		if !pckgNameRegexp.MatchString(pckg) {
			fmt.Println(
				color.RedString(
					"The package name %s is not valid. Please use lowercase latin symbols without spaces",
					pckg,
				),
			)
			continue
		}
		break
	}

	return pckg, nil
}
