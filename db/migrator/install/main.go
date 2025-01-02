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

	err := module.WriteEnvVariablesToFile(envValues, ".env")
	if err != nil {
		fmt.Println("Cannot write env variables to the .env file")
		os.Exit(1)
	}
}
