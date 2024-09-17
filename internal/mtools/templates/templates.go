package templates

import "embed"

//go:embed * */*
var TemplateFiles embed.FS
