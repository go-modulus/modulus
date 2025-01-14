package module

type InstalledFile struct {
	SourceUrl string `json:"sourceUrl"`
	DestFile  string `json:"destFile"`
}

type PostInstallCommand struct {
	CmdPackage string   `json:"cmdPackage"`
	Params     []string `json:"params"`
}

type InstallManifest struct {
	EnvVars             []ConfigEnvVariable  `json:"envVars,omitempty"`
	Dependencies        []string             `json:"dependencies,omitempty"`
	Files               []InstalledFile      `json:"files,omitempty"`
	PostInstallCommands []PostInstallCommand `json:"postInstallCommands,omitempty"`
}

func (m *InstallManifest) AppendEnvVars(vars ...ConfigEnvVariable) *InstallManifest {
	m.EnvVars = append(m.EnvVars, vars...)
	return m
}

func (m *InstallManifest) AppendDependencies(dependencies ...string) *InstallManifest {
	m.Dependencies = append(m.Dependencies, dependencies...)
	return m
}

func (m *InstallManifest) AppendFiles(files ...InstalledFile) *InstallManifest {
	m.Files = append(m.Files, files...)
	return m
}

func (m *InstallManifest) AppendPostInstallCommands(commands ...PostInstallCommand) *InstallManifest {
	m.PostInstallCommands = append(m.PostInstallCommands, commands...)
	return m
}
