package files

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strconv"
)

func AddImportToTools(packageName string) error {
	filename := "tools.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err := os.WriteFile(
			filename,
			[]byte("//go:build tools\n// +build tools\n\npackage tools\n\nimport _ \""+packageName+"\"\n\n"),
			0644,
		)
		if err != nil {
			return err
		}
		return nil
	}
	return AddImportToGoFile(packageName, "_", "tools.go")
}

// AddImportToGoFile add an import package call to a go file
func AddImportToGoFile(
	packageName string,
	alias string,
	filename string,
) error {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, imp := range astFile.Imports {
		if imp.Path.Value == "\""+packageName+"\"" {
			return nil
		}
	}

	var importName *ast.Ident
	if alias != "" {
		importName = &ast.Ident{Name: alias}
	}
	importSpec := &ast.ImportSpec{
		Doc:     nil,
		Name:    importName,
		Path:    &ast.BasicLit{Value: strconv.Quote(packageName), Kind: token.STRING},
		Comment: nil,
		EndPos:  0,
	}

	importDecl := &ast.GenDecl{
		Doc:    nil,
		TokPos: 0,
		Tok:    token.IMPORT,
		Lparen: 0,
		Specs:  []ast.Spec{importSpec},
		Rparen: 0,
	}
	astFile.Decls = append(
		astFile.Decls,
		importDecl,
	)

	ast.SortImports(fset, astFile)
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, astFile); err != nil {
		return err
	}
	return os.WriteFile(filename, buffer.Bytes(), 0644)
}
