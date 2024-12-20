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

func DirExists(dirName string) bool {
	info, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func CreateDirIfNotExists(dirName string) error {
	if DirExists(dirName) {
		return nil
	}
	return os.Mkdir(dirName, 0755)
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

func CopyMakeFileFromTemplates(projPath, srcTmplPath, destName string) error {
	err := CreateDirIfNotExists(projPath + "/mk")
	if err != nil {
		return err
	}
	return CopyFromTemplates(srcTmplPath, projPath+"/mk/"+destName)
}
