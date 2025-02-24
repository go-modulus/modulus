package module

import (
	"context"
	"fmt"
	"github.com/fatih/structs"
	"github.com/sethvargo/go-envconfig"
	"os"
	"sort"
	"strings"
)

type ConfigEnvVariable struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Comment string `json:"comment"`
}

func (v *ConfigEnvVariable) SetComment(comment string) {
	v.Comment = comment
}

func GetEnvVariablesFromConfig[T any](config T) []ConfigEnvVariable {
	vars := getVariables[T](config, true)
	var envVars []ConfigEnvVariable
	for _, value := range vars {
		envVars = append(envVars, value)
	}
	sort.Slice(
		envVars, func(i, j int) bool {
			return envVars[i].Key < envVars[j].Key
		},
	)
	return envVars
}

func WriteEnvVariablesToFile(
	envVars []ConfigEnvVariable,
	filePath string,
) error {
	envVarsMap := make(map[string]ConfigEnvVariable)
	for _, envVar := range envVars {
		envVarsMap[envVar.Key] = envVar
	}
	envFileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(envFileContent), "\n")
	for _, line := range lines {
		if strings.Contains(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		delete(envVarsMap, key)
	}
	if len(envVarsMap) > 0 {
		lines = append(lines, "")

		for _, v := range envVars {
			if envVar, ok := envVarsMap[v.Key]; ok {
				if envVar.Comment != "" {
					lines = append(lines, fmt.Sprintf("# %s", envVar.Comment))
				}
				lines = append(lines, fmt.Sprintf("%s=%s", envVar.Key, envVar.Value))
			}
		}
		lines = append(lines, "")
	}
	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

func getVariables[T any](config T, initDefaults bool) map[string]ConfigEnvVariable {
	envVariables := make(map[string]ConfigEnvVariable)
	if !structs.IsStruct(config) {
		return envVariables
	}
	if initDefaults {
		err := envconfig.Process(context.Background(), &config)
		if err != nil {
			fmt.Println("Error processing config: ", err)
			return envVariables
		}
	}
	fields := structs.Fields(config)
	for _, field := range fields {
		tag := field.Tag("env")
		fieldValue := field.Value()
		tagParts := strings.Split(tag, ",")
		prefix := ""
		comment := field.Tag("comment")
		fieldName := strings.TrimSpace(tagParts[0])
		if len(tagParts) > 1 {
			for _, part := range tagParts {
				optParts := strings.Split(part, "=")
				if len(optParts) < 2 {
					continue
				}
				optKey := strings.TrimSpace(optParts[0])
				if optKey == "prefix" {
					prefix = strings.TrimSpace(optParts[1])
				}
				if optKey == "comment" {
					comment = strings.TrimSpace(optParts[1])
				}
			}
		}
		if structs.IsStruct(fieldValue) {
			subFields := getVariables(fieldValue, false)
			for subFieldName, subFieldValue := range subFields {
				envVariables[prefix+subFieldName] = ConfigEnvVariable{
					Key:     prefix + subFieldName,
					Value:   subFieldValue.Value,
					Comment: subFieldValue.Comment,
				}
			}
		} else {
			envVariables[fieldName] = ConfigEnvVariable{
				Key:     fieldName,
				Value:   fmt.Sprint(fieldValue),
				Comment: comment,
			}
		}
	}

	return envVariables
}
