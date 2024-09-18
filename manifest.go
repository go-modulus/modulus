package modulus

import "embed"

// ManifestFs is the embedded filesystem for the modules manifest files
//
//go:embed modules.json
var ManifestFs embed.FS
