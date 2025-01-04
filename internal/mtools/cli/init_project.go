package cli

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/internal/mtools/utils"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type InitProject struct {
	logger *slog.Logger
}

func NewInitProject(
	logger *slog.Logger,
) *InitProject {
	return &InitProject{
		logger: logger,
	}
}

func NewInitProjectCommand(c *InitProject) *cli.Command {
	return &cli.Command{
		Name: "init",
		Usage: `Inits a project with the base Modulus structure.
Uses interactive prompts to create the project.
Example: ./bin/modulus init --path /path/to/project --name my_project
`,
		Action: c.Invoke,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "Path to the project",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "Name of the project. Will be used as a module name",
			},
		},
	}
}

func (c *InitProject) Invoke(
	ctx *cli.Context,
) error {
	utils.PrintLogo()

	name, path, err := c.getParams(ctx)
	if err != nil {
		fmt.Printf("Error getting the parameters: %s\n", color.RedString(err.Error()))
		return err
	}
	fmt.Println("Start initializing a project")

	err = c.walkToProjectFolder(path)
	if err != nil {
		fmt.Printf("Error during creation the folder structure: %s\n", color.RedString(err.Error()))
		return err
	}

	err = c.createProjectRelatedFiles()
	if err != nil {
		fmt.Printf("Error creating project related files: %s\n", color.RedString(err.Error()))
		return err
	}

	err = c.initGoModules(context.Background(), name)
	if err != nil {
		fmt.Printf("Error initializing the go modules: %s\n", color.RedString(err.Error()))
		return err
	}

	fmt.Println(
		"Congratulations! Your project has been initialized. Please, add your first module.",
	)
	fmt.Println("To add a module, run the command: " + color.CyanString("mtools module install"))

	return nil
}

func (c *InitProject) getParams(ctx *cli.Context) (name, path string, err error) {
	path = ctx.String("path")
	name = ctx.String("name")
	if name == "" {
		name, err = c.askName()
		if err != nil {
			return
		}
	}
	if path == "" {
		path, err = c.askPath(name)
		if err != nil {
			return
		}
	}
	return
}

func (c *InitProject) askName() (string, error) {
	prompt := promptui.Prompt{
		Label: "What is the name of your project?: ",
	}

	return prompt.Run()
}

func (c *InitProject) askPath(name string) (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter a folder : ",
	}

	suggestion := c.getDefaultPath(name)
	prompt.Default = suggestion

	return prompt.Run()
}

func (c *InitProject) getDefaultPath(name string) string {
	nameParts := strings.Split(name, "/")
	return "./" + nameParts[len(nameParts)-1]
}

func (c *InitProject) walkToProjectFolder(path string) error {
	fmt.Println("Creating project folder")
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path+"/internal", 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path+"/cmd/console", 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path+"/bin", 0755)
	if err != nil {
		return err
	}

	fmt.Println("Walking to project folder")
	err = os.Chdir(path)
	if err != nil {
		return err
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println("Changing the folder to " + dir)

	return nil
}

func (c *InitProject) initGoModules(ctx context.Context, name string) error {
	fmt.Println("Initializing go modules")
	if !utils.FileExists("go.mod") {
		err := exec.CommandContext(ctx, "go", "mod", "init", name).Run()
		if err != nil {
			return err
		}
	}
	err := exec.CommandContext(ctx, "go", "get", "github.com/vektra/mockery/v2").Run()
	if err != nil {
		return err
	}

	err = exec.CommandContext(ctx, "go", "get", "github.com/rakyll/gotest").Run()
	if err != nil {
		return err
	}

	err = exec.CommandContext(ctx, "go", "get", "github.com/go-modulus/modulus@latest").Run()
	if err != nil {
		return err
	}

	err = exec.CommandContext(ctx, "go", "get", "-u", "all").Run()
	if err != nil {
		return err
	}
	err = exec.CommandContext(ctx, "go", "mod", "tidy").Run()
	if err != nil {
		return err
	}

	return nil
}

func (c *InitProject) createProjectRelatedFiles() error {
	fmt.Println("Creating project related files")
	names := map[string]string{
		".env":           ".env",
		".env.local":     ".env.local",
		".env.test":      ".env.test",
		"Makefile":       "Makefile",
		"gitignore":      ".gitignore",
		".golangci.yaml": ".golangci.yaml",
		".mockery.yaml":  ".mockery.yaml",
		"tools.go":       "tools.go",
		"main.go":        "cmd/console/main.go",
		"modules.json":   "modules.json",
	}
	for source, name := range names {
		err := utils.CopyFromTemplates("init/"+source, name)
		if err != nil {
			return err
		}
	}

	return nil
}
