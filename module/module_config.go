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
	Key     string
	Value   string
	Comment string
}

func (v *ConfigEnvVariable) SetComment(comment string) {
	v.Comment = comment
}

func GetEnvVariablesFromConfig[T any](config T) []ConfigEnvVariable {
	vars := getVariables[T](config, true)
	var envVars []ConfigEnvVariable
	for key, value := range vars {
		envVars = append(envVars, ConfigEnvVariable{Key: key, Value: value})
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

func getVariables[T any](config T, initDefaults bool) map[string]string {
	envVariables := make(map[string]string)
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
		if structs.IsStruct(fieldValue) {
			tagParts := strings.Split(tag, ",")
			prefix := ""
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
				}
			}
			//fieldName := tagParts[0]
			subFields := getVariables(fieldValue, false)
			for subFieldName, subFieldValue := range subFields {
				envVariables[prefix+subFieldName] = subFieldValue
			}
		} else {
			tagParts := strings.Split(tag, ",")
			fieldName := strings.TrimSpace(tagParts[0])
			envVariables[fieldName] = fmt.Sprint(fieldValue)
		}
	}

	return envVariables
}