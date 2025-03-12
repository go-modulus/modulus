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
	"os"
	"os/exec"
	"text/template"
)

type StorageConfig struct {
	Schema             string
	GenerateGraphql    bool
	GenerateFixture    bool
	GenerateDataloader bool
	ProjPath           string
}

type InstallStorageTmplVars struct {
	Config      StorageConfig
	Module      module.ManifestModule
	StoragePath string
}

type InstallStorage struct {
	UpdateSqlcConfig *UpdateSqlcConfig
}

func NewInstallStorage(config *UpdateSqlcConfig) *InstallStorage {
	return &InstallStorage{
		UpdateSqlcConfig: config,
	}
}

func (c *InstallStorage) Install(ctx context.Context, md module.ManifestModule, cfg StorageConfig) error {
	storagePath := md.StoragePath(cfg.ProjPath)
	err := utils.CreateDirIfNotExists(storagePath)
	if err != nil {
		return fmt.Errorf("cannot create storage directory: %v", err)
	}
	err = utils.CreateDirIfNotExists(storagePath + "/migration")
	if err != nil {
		return fmt.Errorf("cannot create migration directory: %v", err)
	}
	err = utils.CreateDirIfNotExists(storagePath + "/query")
	if err != nil {
		return fmt.Errorf("cannot create query directory: %v", err)
	}

	err = utils.CopyFromTemplates("create_module/sqlc.definition.yaml", cfg.ProjPath+"/sqlc.definition.yaml")
	if err != nil {
		return err
	}

	err = utils.CopyFromTemplates("create_module/default_schema.sql", storagePath+"/migration/default_schema.sql")
	if err != nil {
		return err
	}

	err = utils.CopyFromTemplates("create_module/default_query.sql", storagePath+"/query/default_query.sql")
	if err != nil {
		return err
	}

	vars := InstallStorageTmplVars{
		Config:      cfg,
		Module:      md,
		StoragePath: storagePath,
	}
	err = c.addFilesOfModule(vars, storagePath, cfg.ProjPath)
	if err != nil {
		return err
	}
	err = utils.CopyMakeFileFromTemplates(cfg.ProjPath, "create_module/db.mk", "db.mk")
	if err != nil {
		return err
	}

	// work with sqlc
	err = exec.CommandContext(ctx, "go", "install", "github.com/sqlc-dev/sqlc/cmd/sqlc@latest").Run()
	if err != nil {
		return err
	}
	sqlcFile := storagePath + "/sqlc.yaml"
	err = exec.CommandContext(ctx, "sqlc", "-f", sqlcFile, "generate").Run()
	if err != nil {
		return err
	}
	return nil
}

func (c *InstallStorage) addFilesOfModule(
	vars InstallStorageTmplVars,
	storagePath string,
	projPath string,
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

	err = os.WriteFile(vars.StoragePath+"/sqlc.tmpl.yaml", b.Bytes(), 0644)
	if err != nil {
		fmt.Println(color.RedString("Cannot write a storage tmpl file: %s", err.Error()))
		return err
	}

	err = c.UpdateSqlcConfig.Update(context.Background(), storagePath, projPath)
	if err != nil {
		fmt.Println(color.RedString("Cannot update sqlc config: %s: %s", errors.Hint(err), errors.CauseString(err)))
		return err
	}
	return nil
}
