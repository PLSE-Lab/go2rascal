package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func processFile(filePath string, addLocs bool) string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return fmt.Sprintf("errorFile(Could not process file %s, %s)", filePath, err.Error())
	} else {
		return visitFile(file, fset, addLocs)
	}
}

func computeLocation(fset *token.FileSet, start token.Pos, end token.Pos) string {
	if start.IsValid() && end.IsValid() {
		startPos := fset.Position(start)
		endPos := fset.Position(end)
		return fmt.Sprintf("|file://%s|(%d,%d,<%d,%d>,<%d,%d>)",
			startPos.Filename, startPos.Offset, end-start,
			startPos.Line, startPos.Column-1, endPos.Line, endPos.Column-1)
	} else {
		return fmt.Sprintf("|file://%s|", fset.Position(start).Filename)
	}
}

func visitFile(node *ast.File, fset *token.FileSet, addLocs bool) string {
	var decls []string
	for i := 0; i < len(node.Decls); i++ {
		decls = append(decls, visitDeclaration(&node.Decls[i], fset, addLocs))
	}
	declString := strings.Join(decls, ",")

	var imports []string
	for i := 0; i < len(node.Imports); i++ {
		imports = append(imports, visitImportSpec(node.Imports[i], fset, addLocs))
	}
	importString := strings.Join(imports, ",")

	packageName := node.Name.Name

	if addLocs {
		locationString := computeLocation(fset, node.FileStart, node.FileEnd)
		return fmt.Sprintf("file(\"%s\", [%s], [%s], at=%s)", packageName, declString, importString, locationString)
	} else {
		return fmt.Sprintf("file(\"%s\", [%s], [%s])", packageName, declString, importString)
	}

}

func visitDeclaration(node *ast.Decl, fset *token.FileSet, addLocs bool) string {
	switch d := (*node).(type) {
	case *ast.GenDecl:
		return visitGeneralDeclaration(d, fset, addLocs)
	case *ast.FuncDecl:
		return visitFunctionDeclaration(d, fset, addLocs)
	default:
		return "" // This is an error, should panic here
	}
}

func visitGeneralDeclaration(node *ast.GenDecl, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderDecl(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderDecl()")
	}
}

func visitFunctionDeclaration(node *ast.FuncDecl, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderDecl(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderDecl()")
	}
}

func visitImportSpec(node *ast.ImportSpec, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderImportSpec(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderImportSpec()")
	}
}

func main() {
	var filePath string
	flag.StringVar(&filePath, "filePath", "", "The file to be processed")

	var addLocations bool
	flag.BoolVar(&addLocations, "addLocs", true, "Include location annotations")

	flag.Parse()

	if filePath != "" {
		//fmt.Printf("Processing file %s\n", filePath)
		fmt.Println(processFile(filePath, addLocations))
	} else {
		fmt.Println("No file given")
	}
}
