package config

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/subosito/gotenv"
	"os"
	// strange import. Translation is not working with this import
	_ "golang.org/x/text/message"
)

func LoadDefaultEnv() {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	configDir := os.Getenv("CONFIG_DIR")
	if configDir != "" {
		currentDir = configDir
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}
	// load ".env.{APP_ENV}". This is the .env file with overrides of default variables.
	// It is loaded first so that it can override the default variables
	LoadEnv(currentDir, env, false)
	// load default ".env". This is the .env file with default variables.
	LoadEnv(currentDir, "", false)
}

func IsProd() bool {
	return os.Getenv("APP_ENV") == "prod"
}

func LoadEnv(basePath string, env string, override bool) {
	if env != "" {
		env = "." + env
	}

	success := load(
		[]string{
			basePath + "/.env" + env,
		}, override,
	)

	if ok := os.Getenv("DEBUG"); ok != "" {
		if success {
			fmt.Println("Config is loaded from", color.BlueString(basePath+"/.env"+env))
		} else {
			fmt.Println("Config is not found ", color.RedString(basePath+"/.env"+env))
		}
	}
}

func load(filenames []string, override bool) bool {
	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return false
			}
			panic(err)
		}
		if override {
			err = gotenv.OverApply(f)
		} else {
			err = gotenv.Apply(f)
		}
		f.Close()
		if err != nil {
			panic(err)
		}
	}

	return true
}
