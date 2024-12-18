package action

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/internal/mtools/templates"
	"github.com/go-modulus/modulus/internal/mtools/utils"
	"github.com/go-modulus/modulus/module"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"os"
	"text/template"
)

type StorageConfig struct {
	Schema             string
	GenerateGraphql    bool
	GenerateFixture    bool
	GenerateDataloader bool
}

type InstallStorageTmplVars struct {
	Config StorageConfig
	Module module.ManifestItem
}

type InstallStorage struct {
	UpdateSqlcConfig *UpdateSqlcConfig
}

func NewInstallStorage(config *UpdateSqlcConfig) *InstallStorage {
	return &InstallStorage{
		UpdateSqlcConfig: config,
	}
}

func (c *InstallStorage) Install(ctx context.Context, md module.ManifestItem, cfg StorageConfig) error {
	os.Mkdir(md.LocalPath+"/storage", 0755)
	os.Mkdir(md.LocalPath+"/storage/migration", 0755)
	os.Mkdir(md.LocalPath+"/storage/query", 0755)

	err := utils.CopyFromTemplates("create_module/sqlc.definition.yaml", "./sqlc.definition.yaml")
	if err != nil {
		return err
	}

	vars := InstallStorageTmplVars{
		Config: cfg,
		Module: md,
	}
	err = c.addModuleFile(vars)
	if err != nil {
		return err
	}

	return nil
}

func (c *InstallStorage) addModuleFile(
	vars InstallStorageTmplVars,
) error {
	tmpl := template.Must(
		template.New("sqlc.yaml.tmpl").
			ParseFS(
				templates.TemplateFiles,
				"create_module/sqlc.yaml.tmpl",
			),
	)

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := tmpl.ExecuteTemplate(w, "sqlc.yaml.tmpl", &vars)
	if err != nil {
		return err
	}
	w.Flush()

	err = os.WriteFile(vars.Module.LocalPath+"/storage/sqlc.tmpl.yaml", b.Bytes(), 0644)
	if err != nil {
		fmt.Println(color.RedString("Cannot write a storage tmpl file: %s", err.Error()))
		return err
	}

	err = c.UpdateSqlcConfig.Update(context.Background(), vars.Module)
	if err != nil {
		p := message.NewPrinter(language.English)
		fmt.Println(color.RedString("Cannot update sqlc config: %s: %s", errors.Hint(p, err), errors.CauseString(err)))
		return err
	}
	return nil
}
