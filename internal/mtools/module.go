package mtools

import (
	cli2 "github.com/go-modulus/modulus/internal/mtools/cli"
	"github.com/go-modulus/modulus/logger"
	"github.com/go-modulus/modulus/module"
)

func NewModule() *module.Module {
	return module.NewModule("github.com/go-modulus/modulus/mtools").
		AddCliCommand(cli2.NewCommand, cli2.NewInitProject).
		AddDependencies(*logger.NewModule(logger.ModuleConfig{}))
}
