package main

import (
	"fmt"
	"github.com/go-modulus/modulus/db/pgx"
	"github.com/go-modulus/modulus/module"
	"os"
)

func main() {
	moduleConfig := pgx.ModuleConfig{}

	envValues := module.GetEnvVariablesFromConfig(moduleConfig)
	for i, value := range envValues {
		if value.Key == "PGX_DSN" {
			envValues[i].SetComment("Use this variable to set the DSN for the PGX connection. It overwrites the other PG_* variables.")
		}
	}

	err := module.WriteEnvVariablesToFile(envValues, ".env")
	if err != nil {
		fmt.Println("Cannot write env variables to the .env file")
		os.Exit(1)
	}
}
