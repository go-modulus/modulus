package files

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"os"
	"strconv"
	"strings"
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
	_, err := AddImportToGoFile(packageName, "_", "tools.go")
	return err
}

// AddImportToGoFile add an import package call to a go file
// Returns the package name that can be used in calls
func AddImportToGoFile(
	packageName string,
	alias string,
	filename string,
) (string, error) {
	pkgName := alias
	if alias == "" || alias == "_" {
		parts := strings.Split(packageName, "/")
		pkgName = parts[len(parts)-1]
	}
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return pkgName, err
	}

	i := 1
	basePkgName := pkgName
	for {
		breakAfterLoop := true
		for _, imp := range astFile.Imports {
			pkgAlias := ""
			if imp.Name != nil {
				pkgAlias = imp.Name.Name
			}
			if pkgAlias == "" {
				parts := strings.Split(imp.Path.Value, "/")
				pkgAlias = strings.Trim(parts[len(parts)-1], "\"")
			}
			if imp.Path.Value == "\""+packageName+"\"" {
				return pkgAlias, nil
			}
			if pkgName == pkgAlias && pkgName != "_" {
				i++
				pkgName = basePkgName + strconv.Itoa(i)
				alias = pkgName
				breakAfterLoop = false
				break
			}
		}
		if breakAfterLoop {
			break
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
		return alias, err
	}
	return alias, os.WriteFile(filename, buffer.Bytes(), 0644)
}

func AddModuleToEntrypoint(
	packagePath string,
	filename string,
) error {
	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	imports := astutil.Imports(fset, astFile)

	alias, err := getUniqAlias(packagePath, 0, imports)
	if err != nil {
		return err
	}
	if alias == getDefPkgName(packagePath) {
		astutil.AddImport(fset, astFile, packagePath)
	} else {
		astutil.AddNamedImport(fset, astFile, alias, packagePath)
	}

	astutil.Apply(astFile, initImportedModule(alias), nil)

	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, astFile); err != nil {
		return err
	}
	source, err := format.Source(buffer.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(filename, source, 0644)
}

func initImportedModule(alias string) astutil.ApplyFunc {
	return func(cursor *astutil.Cursor) bool {
		//add a value to a slice with name s
		if cursor.Name() == "Body" {
			body, ok := cursor.Node().(*ast.BlockStmt)
			if !ok {
				return true
			}
			for _, stmt := range body.List {
				astmt, ok := stmt.(*ast.AssignStmt)
				if !ok {
					continue
				}
				expr, ok := astmt.Lhs[0].(*ast.Ident)
				if !ok {
					continue
				}
				if expr.Name == "importedModulesOptions" {
					arExpr, ok := astmt.Rhs[0].(*ast.CompositeLit)
					if !ok {
						continue
					}
					if isImportInitialized(alias, arExpr) {
						break
					}
					arExpr.Elts = append(
						arExpr.Elts,
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: "\n" + alias + ".NewModule().BuildFx(),\n",
						},
					)
					break
				}
			}
		}
		return true
	}
}

func isImportInitialized(alias string, arExpr *ast.CompositeLit) bool {
	for _, elt := range arExpr.Elts {
		v, ok := elt.(*ast.CallExpr)
		if !ok {
			continue
		}
		buildFxExpr, ok := v.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if buildFxExpr.Sel.Name != "BuildFx" {
			continue
		}

		newModuleExprX, ok := buildFxExpr.X.(*ast.CallExpr)
		if !ok {
			continue
		}
		newModuleExpr, ok := newModuleExprX.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if newModuleExpr.Sel.Name != "NewModule" {
			continue
		}

		ident, ok := newModuleExpr.X.(*ast.Ident)
		if !ok {
			continue
		}

		if ident.Name == alias {
			return true
		}
	}

	return false
}

func getDefPkgName(packagePath string) string {
	parts := strings.Split(packagePath, "/")
	return strings.Trim(parts[len(parts)-1], "\"")
}

func getUniqAlias(
	packagePath string,
	aliasIterator int,
	imports [][]*ast.ImportSpec,
) (alias string, err error) {
	alias = getDefPkgName(packagePath)
	// need to add a number to the alias starting from 2 if the default alias is already used
	if aliasIterator < 2 {
		aliasIterator++
	}
	if aliasIterator > 1 {
		alias += strconv.Itoa(aliasIterator)
	}

	for _, importSpecs := range imports {
		for _, imp := range importSpecs {
			pkgAlias := ""
			if imp.Name != nil {
				pkgAlias = imp.Name.Name
			}
			if pkgAlias == "" {
				pkgAlias = getDefPkgName(imp.Path.Value)
			}
			if imp.Path.Value == "\""+packagePath+"\"" {
				return pkgAlias, nil
			}
			if alias == pkgAlias {
				return getUniqAlias(packagePath, aliasIterator+1, imports)
			}
		}
	}
	return alias, nil
}
