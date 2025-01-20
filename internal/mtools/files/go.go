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

	if alias == "" {
		astutil.AddImport(fset, astFile, packageName)
	} else {
		astutil.AddNamedImport(fset, astFile, alias, packageName)
	}

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

	for _, specs := range imports {
		for _, spec := range specs {
			if spec.Path.Value == "\""+packagePath+"\"" {
				return nil
			}
		}
	}

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
				if expr.Name == "modules" {
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
							Value: "\n" + alias + ".NewModule(),\n",
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

func AddConstructorToProvider(
	packagePath string,
	constructor string,
	filename string,
) error {
	return addConstructor(packagePath, constructor, filename, "AddProviders")
}

func AddCliCommand(
	packagePath string,
	constructor string,
	filename string,
) error {
	return addConstructor(packagePath, constructor, filename, "AddCliCommands")
}

func addConstructor(
	packagePath string,
	constructor string,
	filename string,
	extendedMethodName string,
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

	//astFile.
	astutil.Apply(astFile, addProvider(alias, constructor, extendedMethodName), nil)

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

func addProvider(alias string, constructor string, extendedMethodName string) astutil.ApplyFunc {
	return func(cursor *astutil.Cursor) bool {
		//add a value to a slice with name s
		if cursor.Name() == "Body" {
			body, ok := cursor.Node().(*ast.BlockStmt)
			if !ok {
				return true
			}
			//check if the new module is returned from function
			// func NewModule() *Module {
			// 	return module.NewModule().AddProvider()
			// }
			for _, stmt := range body.List {
				rstmt, ok := stmt.(*ast.ReturnStmt)
				if !ok {
					continue
				}
				if len(rstmt.Results) == 1 {
					nextNode := rstmt.Results[0]
					if injectConstructorToAddProviders(nextNode, alias, constructor, extendedMethodName) {
						return false
					}
				}
			}

			// check if the new module is saved into the variable
			// m := module.NewModule().AddProvider()
			// return m
			for _, stmt := range body.List {
				astmt, ok := stmt.(*ast.AssignStmt)
				if !ok {
					continue
				}

				nextNode := astmt.Rhs[0]
				if injectConstructorToAddProviders(nextNode, alias, constructor, extendedMethodName) {
					return false
				}
			}
		}
		return true
	}
}

func injectConstructorToAddProviders(
	nextNode ast.Expr,
	alias string,
	constructor string,
	extendedMethodName string,
) bool {
	for {
		callExpr, ok := nextNode.(*ast.CallExpr)
		if !ok {
			return false
		}
		selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		if selectorExpr.Sel.Name == extendedMethodName {
			position := callExpr.Rparen
			callExpr.Args = append(
				callExpr.Args,
				&ast.SelectorExpr{
					X:   &ast.Ident{Name: alias, NamePos: position},
					Sel: &ast.Ident{Name: constructor + ",\n", NamePos: position},
				},
			)

			return true
		}
		nextNode = selectorExpr.X
	}
}
