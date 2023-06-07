package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
)

// parses file and makes sure there is no error in file input.
func processFile(filePath string, addLocs bool) string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return fmt.Sprintf("errorFile(Could not process file %s, %s)", filePath, err.Error())
	} else {
		return visitFile(file, fset, addLocs)
	}
}

// computes location of the ast nodes
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

// ------------------------------------------------------------------------------------------------------------------
// start of ast vistors
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
		return fmt.Sprintf("placeholderDecl(at=%s), %s", locationString, vistFunctionBody(node.Body, fset, addLocs))
	} else {
		return fmt.Sprintf("placeholderDecl(), %s", vistFunctionBody(node.Body, fset, addLocs))
	}
}

//end of declarations.
//---------------------------------------------------------------------------------------------------------------------------

func visitImportSpec(node *ast.ImportSpec, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderImportSpec(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderImportSpec()")
	}
}

// end of imports.
// ----------------------------------------------------------------------------------------------------------------------------

func vistFunctionBody(node *ast.BlockStmt, fset *token.FileSet, addLocs bool) string {
	locationString := computeLocation(fset, node.Pos(), node.End())
	var stmts []string
	for i := 0; i < len(node.List); i++ {
		stmts = append(stmts, visitStmt(&node.List[i], fset, addLocs))
	}

	stmtsString := strings.Join(stmts, ",")
	if addLocs {
		return fmt.Sprintf("placeholderBody(at=%s), %s", locationString, stmtsString)
	} else {
		return fmt.Sprintf("placeholderBody(), %s", stmtsString)
	}
}

func visitStmt(node *ast.Stmt, fset *token.FileSet, addLocs bool) string {
	switch t := (*node).(type) {
	case *ast.DeclStmt:
		return visitDeclStmt(t, fset, addLocs)
	case *ast.EmptyStmt:
		return visitEmptyStmt(t, fset, addLocs)
	case *ast.LabeledStmt:
		return visitLabeledStmt(t, fset, addLocs)
	case *ast.ExprStmt:
		return visitExprStmt(t, fset, addLocs)
	case *ast.SendStmt:
		return visitSendStmt(t, fset, addLocs)
	case *ast.IncDecStmt:
		return visitIncDecStmt(t, fset, addLocs)
	case *ast.AssignStmt:
		return visitAssignStmt(t, fset, addLocs)
	case *ast.GoStmt:
		return visitGoStmt(t, fset, addLocs)
	case *ast.DeferStmt:
		return visitDeferStmt(t, fset, addLocs)
	case *ast.ReturnStmt:
		return visitReturnStmt(t, fset, addLocs)
	case *ast.BranchStmt:
		return visitBranchStmt(t, fset, addLocs)
	case *ast.BlockStmt:
		return visitBlockStmt(t, fset, addLocs)
	case *ast.IfStmt:
		return visitIfStmt(t, fset, addLocs)
	case *ast.CaseClause:
		return visitCaseClause(t, fset, addLocs)
	case *ast.SwitchStmt:
		return visitSwitchStmt(t, fset, addLocs)
	case *ast.TypeSwitchStmt:
		return visitTypeSwitchStmt(t, fset, addLocs)
	case *ast.CommClause:
		return visitCommClause(t, fset, addLocs)
	case *ast.SelectStmt:
		return visitSelectStmt(t, fset, addLocs)
	case *ast.ForStmt:
		return visitForStmt(t, fset, addLocs)
	case *ast.RangeStmt:
		return visitRangeStmt(t, fset, addLocs)
	}
	return reflect.TypeOf(*node).Elem().Name()
}

func visitDeclStmt(node *ast.DeclStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholdeDeclStmt(%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderDeclStmt()")
	}
}

func visitEmptyStmt(node *ast.EmptyStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderEmptyStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderEmptyStmt()")
	}
}

func visitLabeledStmt(node *ast.LabeledStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderLabeledStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderLabeledStmt()")
	}
}

func visitExprStmt(node *ast.ExprStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderExprStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderExprStmt()")
	}
}

func visitSendStmt(node *ast.SendStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderSendStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderSendStmt()")
	}
}

func visitIncDecStmt(node *ast.IncDecStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderIncDecStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderIncDecStmt()")
	}
}

func visitAssignStmt(node *ast.AssignStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderAssignStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderAssignStmt()")
	}
}

func visitGoStmt(node *ast.GoStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderGoStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderGoStmt()")
	}
}

func visitDeferStmt(node *ast.DeferStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderDeferStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderDeferStmt()")
	}
}

func visitReturnStmt(node *ast.ReturnStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderReturnStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderReturnStmt()")
	}
}

func visitBranchStmt(node *ast.BranchStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderBranchStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderBranchStmt()")
	}
}

func visitBlockStmt(node *ast.BlockStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderBlockStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderBlockStmt()")
	}
}

func visitIfStmt(node *ast.IfStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderIfStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderIfStmt()")
	}
}

func visitCaseClause(node *ast.CaseClause, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderCaseClause(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderCaseClause()")
	}
}

func visitSwitchStmt(node *ast.SwitchStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderSwitchStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderSwitchStmt()")
	}
}

func visitTypeSwitchStmt(node *ast.TypeSwitchStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderTypeSwitchStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderTypeSwitchStmt()")
	}
}

func visitCommClause(node *ast.CommClause, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderCommClause(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderCommClause()")
	}
}

func visitSelectStmt(node *ast.SelectStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderSelectStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderSelectStmt()")
	}
}

func visitForStmt(node *ast.ForStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderForStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderForStmt()")
	}
}

func visitRangeStmt(node *ast.RangeStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderRangeStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderRangeStmt()")
	}
}

//end of statements.
//------------------------------------------------------------------------------------------------------------------------------------------

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
