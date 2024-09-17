package utils

import (
	"github.com/go-modulus/modulus/internal/mtools/templates"
	"os"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func CopyFromTemplates(src, dest string) error {
	if FileExists(dest) {
		return nil
	}
	content, err := templates.TemplateFiles.ReadFile(src)
	if err != nil {
		return err
	}
	err = os.WriteFile(dest, content, 0644)
	if err != nil {
		return err
	}
	return nil
}
