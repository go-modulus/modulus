package config

import (
	"errors"
	"github.com/subosito/gotenv"
	"os"
	// strange import. Translation is not working with this import
	_ "golang.org/x/text/message"
)

func LoadDefaultEnv() {
	LoadEnv("", "", false)
	LoadEnv("", os.Getenv("APP_ENV"), true)
}

func IsProd() bool {
	return os.Getenv("APP_ENV") == "prod"
}

func LoadEnv(basePath string, env string, override bool) {
	if env != "" {
		env = "." + env
	}
	load(
		[]string{
			basePath + "/.env" + env,
			basePath + "/.env" + env + ".local",
		}, override,
	)
}

func load(filenames []string, override bool) {
	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return
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
}
