package module

type InstallManifest struct {
	EnvVars      []ConfigEnvVariable `json:"envVars,omitempty"`
	Dependencies []string            `json:"dependencies,omitempty"`
}

func (m *InstallManifest) AppendEnvVars(vars ...ConfigEnvVariable) {
	m.EnvVars = append(m.EnvVars, vars...)
}

func (m *InstallManifest) AppendDependencies(dependencies ...string) {
	m.Dependencies = append(m.Dependencies, dependencies...)
}
