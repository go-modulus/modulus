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
	"net/http"
	"regexp"
	"strings"
)

var apiHandlerNameRegEx = regexp.MustCompile(`^[A-Z]+[a-zA-Z0-9_]*$`)

type AddJsonApiTmplVars struct {
	StructName  string
	Uri         string
	PackageName string
	Method      string
}

func (v AddJsonApiTmplVars) IsBodyRequired() bool {
	return v.Method == http.MethodPost || v.Method == http.MethodPut
}

type AddJsonApi struct {
}

func NewAddJsonApi() *AddJsonApi {
	return &AddJsonApi{}
}
func NewAddJsonApiCommand(addJsonApi *AddJsonApi) *cli.Command {
	return &cli.Command{
		Name: "add-json-api",
		Usage: `Add a boilerplate for the API handler that uses JSON as a transport.
Example: mtools module add-json-api
Example: mtools module add-json-api --module=example --uri=/hello-world --name=HelloWorld --method=GET --silent
`,
		Action: addJsonApi.Invoke,
		Flags: []cli.Flag{
			flag.NewModule("A module name to add a API handler to"),
			&cli.StringFlag{
				Name:    "name",
				Usage:   "The struct name of the created API handler",
				Aliases: []string{"n"},
			},
			&cli.StringFlag{
				Name:    "uri",
				Usage:   "The URI of the API handler",
				Aliases: []string{"u"},
			},
			&cli.StringFlag{
				Name:    "method",
				Usage:   "HTTP method for the API handler",
				Aliases: []string{"mt"},
			},
			flag.NewSilent("Do not ask for any input"),
		},
	}
}

func (a *AddJsonApi) Invoke(ctx *cli.Context) error {
	mod, err := flag.ModuleValue(ctx)
	if err != nil {
		return nil
	}
	isSilent := flag.SilentValue(ctx)
	projPath := flag.ProjPathValue(ctx)

	name := ctx.String("name")
	if name == "" {
		if isSilent {
			fmt.Println(color.RedString("The API handler name is required"))
			return nil
		}
		name = a.askApiHandlerName()
		if name == "" {
			return nil
		}
	}

	method := ctx.String("method")
	if method == "" {
		if isSilent {
			method = http.MethodGet
		} else {
			method = a.askApiHandlerMethod()
		}
	}

	fmt.Println(
		color.GreenString("Adding an HTTP API handler %s"),
		color.BlueString(name),
		color.GreenString(
			"to the module %s",
			color.BlueString(mod.Name),
		),
	)

	structName := strcase.ToCamel(name)

	uri := ctx.String("uri")
	if uri == "" {
		if isSilent {
			fmt.Println(color.RedString("The API handler URI is required"))
			return nil
		}
		uri = a.askApiHandlerUri()
		if uri == "" {
			return nil
		}
	}

	err = utils.CreateDirIfNotExists(mod.ApiPath(projPath))
	if err != nil {
		fmt.Println(
			color.RedString("Cannot create the API directory"),
			color.BlueString(mod.ApiPath(projPath)),
			color.RedString(": %s", err.Error()),
		)
		return nil
	}

	err = a.createApiHandlerFile(structName, uri, mod, projPath, method)
	if err != nil {
		return nil
	}

	return nil
}

func (a *AddJsonApi) createApiHandlerFile(
	structName string,
	uri string,
	mod module.ManifestModule,
	projPath string,
	method string,
) error {
	path := mod.ApiPath(projPath)
	pckg := mod.ApiPackage()
	tmplVars := AddJsonApiTmplVars{
		StructName:  structName,
		Uri:         uri,
		PackageName: "api",
		Method:      method,
	}

	handlerFile := strcase.ToSnake(structName)

	err := utils.ProcessTemplate(
		"api_handler.go.tmpl",
		"add_json_api/api_handler.go.tmpl",
		path+"/"+handlerFile+".go",
		tmplVars,
	)
	if err != nil {
		fmt.Println(
			color.RedString("Cannot create the API handler: %s", err.Error()),
		)
		return err
	}

	moduleFile := mod.ModulePath(projPath) + "/module.go"
	// this call is not necessary, but it's fine to have an import with defined alias
	_, err = files.AddImportToGoFile(pckg, "api", moduleFile)
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

	err = files.AddConstructorToProvider(pckg, "New"+structName+"Route", moduleFile)
	if err != nil {
		fmt.Println(
			color.RedString("Cannot add a route constructor to the module.go file: %s", err.Error()),
		)
		return err
	}

	return nil
}

func (a *AddJsonApi) askApiHandlerName() string {
	for {
		prompt := promptui.Prompt{
			Label: "Enter an Api handler struct name in the CamelCase. Example: HelloWorld",
		}

		name, err := prompt.Run()
		if err != nil {
			fmt.Println(color.RedString("Cannot ask api handler name: %s", err.Error()))
			return ""
		}
		if name == "" {
			fmt.Println(color.RedString("The api handler name cannot be empty"))
			continue
		}
		if !apiHandlerNameRegEx.MatchString(name) {
			fmt.Println(color.RedString("The api handler name must be in the CamelCase. Latin letters, numbers, and underscores are allowed."))
			fmt.Println(color.BlueString("Example: HelloWorld"))
			continue
		}
		return name
	}
}

func (a *AddJsonApi) askApiHandlerUri() string {
	for {
		prompt := promptui.Prompt{
			Label: "Enter an Api handler URI. Example: /hello-world",
		}

		uri, err := prompt.Run()
		if err != nil {
			fmt.Println(color.RedString("Cannot ask api handler uri: %s", err.Error()))
			return ""
		}
		if uri == "" {
			fmt.Println(color.RedString("The api handler URI cannot be empty"))
			continue
		}
		if uri[0] != '/' {
			fmt.Println(color.RedString("The api handler URI must start with a slash"))
			fmt.Println(color.BlueString("Example: /hello-world"))
			continue
		}
		return uri
	}
}

func (a *AddJsonApi) askApiHandlerMethod() string {
	for {
		prompt := promptui.Prompt{
			Label: "Enter an Api handler method. Example: GET",
		}

		method, err := prompt.Run()
		if err != nil {
			fmt.Println(color.RedString("Cannot ask api handler method: %s", err.Error()))
			return ""
		}
		if method == "" {
			fmt.Println(color.RedString("The api handler method cannot be empty"))
			continue
		}
		method = strings.ToUpper(method)
		if method != http.MethodGet && method != http.MethodPost && method != http.MethodPut && method != http.MethodDelete {
			fmt.Println(color.RedString("The api handler method must be one of the following: GET, POST, PUT, DELETE"))
			fmt.Println(color.BlueString("Example: GET"))
			continue
		}
		return method
	}
}
