package action

import (
	"context"
	"errors"
	errors2 "github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errbuilder"
	"github.com/go-modulus/modulus/module"
	"gopkg.in/yaml.v3"
	"os"
)

var ErrSqlcDefinitionFileNotFound = errors.New("project_root/sqlc.definition.yaml file not found")
var ErrSqlcTemplateFileNotFound = errors.New("module_path/storage/sqlc.tmpl.yaml file not found")
var ErrCannotParseSqlcDefinition = errbuilder.New("cannot parse sqlc.definition.yaml file").
	WithHint("Please check the file sqlc.definition.yaml content. It has wrong yaml format.").
	Build()
var ErrNoDefinitionBlock = errbuilder.New("sqlc.definition.yaml file does not have definition block").
	WithHint("Please check the file sqlc.definition.yaml content. It has no the 'definition' map field.").
	Build()
var ErrCannotParseSqlcTmpl = errbuilder.New("cannot parse sqlc.tmpl.yaml file").
	WithHint("Please check the file module/storage/sqlc.tmpl.yaml content. It has wrong yaml format.").
	Build()
var ErrCannotUpdateSqlcConfig = errbuilder.New("cannot update sqlc config").
	WithHint("Some issues occurred when the sql.yaml file is being combined.").
	Build()

type UpdateSqlcConfig struct {
}

func NewUpdateSqlcConfig() *UpdateSqlcConfig {
	return &UpdateSqlcConfig{}
}

func (c *UpdateSqlcConfig) Update(ctx context.Context, md module.ManifestItem) error {
	defFile := "./sqlc.definition.yaml"
	defContent, err := os.ReadFile(defFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrSqlcDefinitionFileNotFound
		}
		return err
	}

	var def map[string]interface{}

	err = yaml.Unmarshal(defContent, &def)
	if err != nil {
		return errors2.WrapCause(ErrCannotParseSqlcDefinition, err)
	}

	tmplContent, err := os.ReadFile(md.LocalPath + "/storage/sqlc.tmpl.yaml")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrSqlcTemplateFileNotFound
		}
		return err
	}

	resContent := defContent
	resContent = append(resContent, []byte("\n\n")...)
	resContent = append(resContent, tmplContent...)

	var tmpl map[string]interface{}

	err = yaml.Unmarshal(resContent, &tmpl)
	if err != nil {
		return errors2.WrapCause(ErrCannotParseSqlcTmpl, err)
	}
	for key, val := range def {
		tmpl[key] = val
	}

	tmplContent, err = yaml.Marshal(tmpl)
	if err != nil {
		return errors2.WrapCause(ErrCannotUpdateSqlcConfig, err)
	}

	sqlcContent, err := yaml.Marshal(tmpl["sqlc-tmpl"])

	err = os.WriteFile(md.LocalPath+"/storage/sqlc.yaml", sqlcContent, 0644)
	if err != nil {
		return errors2.WrapCause(ErrCannotUpdateSqlcConfig, err)
	}

	return nil
}
