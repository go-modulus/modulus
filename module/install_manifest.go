package module

type InstallManifest struct {
	EnvVars []ConfigEnvVariable `json:"envVars,omitempty"`
}

func (m *InstallManifest) AppendEnvVars(vars ...ConfigEnvVariable) {
	m.EnvVars = append(m.EnvVars, vars...)
}
