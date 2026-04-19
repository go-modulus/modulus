package module

type InstalledFile struct {
	SourceUrl string `json:"sourceUrl"`
	DestFile  string `json:"destFile"`
}

type PostInstallCommand struct {
	CmdPackage string   `json:"cmdPackage"`
	Params     []string `json:"params"`
}

type InstallationManifesto struct {
	EnvVars             []ConfigEnvVariable  `json:"envVars,omitempty"`
	Dependencies        []string             `json:"dependencies,omitempty"`
	Files               []InstalledFile      `json:"files,omitempty"`
	PostInstallCommands []PostInstallCommand `json:"postInstallCommands,omitempty"`
}

func (m *InstallationManifesto) AppendEnvVars(vars ...ConfigEnvVariable) *InstallationManifesto {
	m.EnvVars = append(m.EnvVars, vars...)
	return m
}

func (m *InstallationManifesto) AppendDependencies(dependencies ...string) *InstallationManifesto {
	m.Dependencies = append(m.Dependencies, dependencies...)
	return m
}

func (m *InstallationManifesto) AppendFiles(files ...InstalledFile) *InstallationManifesto {
	m.Files = append(m.Files, files...)
	return m
}

func (m *InstallationManifesto) AppendPostInstallCommands(commands ...PostInstallCommand) *InstallationManifesto {
	m.PostInstallCommands = append(m.PostInstallCommands, commands...)
	return m
}
