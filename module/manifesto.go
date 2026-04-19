package module

import (
	"strings"
)

type Manifesto struct {
	Name          string                `json:"name"`
	Package       string                `json:"package"`
	Description   string                `json:"description"`
	Install       InstallationManifesto `json:"install,omitempty"`
	Version       string                `json:"version"`
	LocalPath     string                `json:"localPath,omitempty"`
	IsLocalModule bool                  `json:"isLocalModule,omitempty"`
}

func NewManifesto(
	module *Module,
	pkg string,
	description string,
	version string,
) Manifesto {
	deps := make([]string, 0, len(module.dependencies))
	for _, dep := range module.dependencies {
		deps = append(deps, dep.name)
	}
	install := InstallationManifesto{}
	if len(module.envVars) > 0 {
		install.
			AppendEnvVars(module.envVars...)
	}
	if len(deps) > 0 {
		install.
			AppendDependencies(deps...)
	}

	currentModule := Manifesto{
		Name:        module.name,
		Package:     pkg,
		Install:     install,
		Version:     version,
		Description: description,
	}

	return currentModule
}

type Entrypoint struct {
	LocalPath string `json:"localPath"`
	Name      string `json:"name"`
}

func (m Manifesto) GetShortPackageName() string {
	return m.Package[strings.LastIndex(m.Package, "/")+1:]
}

func (m Manifesto) StoragePath(projPath string) string {
	return m.ModulePath(projPath) + "/storage"
}

func (m Manifesto) ModulePath(projPath string) string {
	if projPath == "" {
		return m.LocalPath
	}
	return projPath + "/" + m.LocalPath
}

func (m Manifesto) StoragePackage() string {
	return m.Package + "/storage"
}

func (m Manifesto) CliPath(projPath string) string {
	return m.ModulePath(projPath) + "/cli"
}

func (m Manifesto) CliPackage() string {
	return m.Package + "/cli"
}

func (m Manifesto) ApiPath(projPath string) string {
	return m.ModulePath(projPath) + "/api"
}

func (m Manifesto) ApiPackage() string {
	return m.Package + "/api"
}
