package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func processFile(filePath string) string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return fmt.Sprintf("errorFile(Could not process file %s, %s)", filePath, err.Error())
	} else {
		return visitFile(file, fset)
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

func visitFile(node *ast.File, fset *token.FileSet) string {
	var decls []string
	for i := 0; i < len(node.Decls); i++ {
		decls = append(decls, visitDeclaration(&node.Decls[i], fset))
	}
	declString := strings.Join(decls, ",")

	var imports []string
	for i := 0; i < len(node.Imports); i++ {
		imports = append(imports, visitImportSpec(node.Imports[i], fset))
	}
	importString := strings.Join(imports, ",")

	locationString := computeLocation(fset, node.FileStart, node.FileEnd)

	packageName := node.Name.Name
	return fmt.Sprintf("file(\"%s\", [%s], [%s], %s)", packageName, declString, importString, locationString)
}

func visitDeclaration(node *ast.Decl, fset *token.FileSet) string {
	//locationString := computeLocation(fset, node.FileStart, node.FileEnd)
	return ""
}

func visitImportSpec(node *ast.ImportSpec, fset *token.FileSet) string {
	//locationString := computeLocation(fset, node.FileStart, node.FileEnd)
	return ""
}

func main() {
	var filePath string
	flag.StringVar(&filePath, "filePath", "", "The file to be processed")

	flag.Parse()

	if filePath != "" {
		fmt.Printf("Processing file %s\n", filePath)
		fmt.Println(processFile(filePath))
	} else {
		fmt.Println("No file given")
	}
}
