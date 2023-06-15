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
		return fmt.Sprintf("data function = (name=%s, at=%s), %s", node.Name, locationString, visitFunctionBody(node.Body, fset, addLocs))
	} else {
		return fmt.Sprintf("data function = (name=%s), %s", node.Name, visitFunctionBody(node.Body, fset, addLocs))
	}
}

//end of declarations.
//---------------------------------------------------------------------------------------------------------------------------

func visitImportSpec(node *ast.ImportSpec, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("data import = (name=%s, at=%s), ", node.Name, locationString)
	} else {
		return fmt.Sprintf("data import = (name=%s), ", node.Name)
	}
}

// end of imports.
// ----------------------------------------------------------------------------------------------------------------------------

func visitFunctionBody(node *ast.BlockStmt, fset *token.FileSet, addLocs bool) string {
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
		return fmt.Sprintf("placeholdeDeclStmt(%s), %s", locationString, visitDeclaration(&node.Decl, fset, addLocs))
	} else {
		return fmt.Sprintf("placeholderDeclStmt(), %s", visitDeclaration(&node.Decl, fset, addLocs))
	}
}

// need to come back to I think
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
		return fmt.Sprintf("placeholderLabeledStmt(at=%s), %s", locationString, visitStmt(&node.Stmt, fset, addLocs))
	} else {
		return fmt.Sprintf("placeholderLabeledStmt(), %s", visitStmt(&node.Stmt, fset, addLocs))
	}
}

//-----------------------------------------------------------------

func visitExprStmt(node *ast.ExprStmt, fset *token.FileSet, addLocs bool) string {
	e := visitExpr(&node.X, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderExprStmt(at=%s), %s", locationString, e)
	} else {
		return fmt.Sprintf("placeholderExprStmt() %s", e)
	}
}

func visitSendStmt(node *ast.SendStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderSendStmt(at=%s), %s, %s", locationString, visitExpr(&node.Chan, fset, addLocs), visitExpr(&node.Value, fset, addLocs))
	} else {
		return fmt.Sprintf("placeholderSendStmt(), %s, %s", visitExpr(&node.Chan, fset, addLocs), visitExpr(&node.Value, fset, addLocs))
	}
}

func visitIncDecStmt(node *ast.IncDecStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderIncDecStmt(at=%s), %s", locationString, visitExpr(&node.X, fset, addLocs))
	} else {
		return fmt.Sprintf("placeholderIncDecStmt(), %s", visitExpr(&node.X, fset, addLocs))
	}
}

func visitAssignStmt(node *ast.AssignStmt, fset *token.FileSet, addLocs bool) string {
	right := visitExprList(node.Rhs, fset, addLocs)
	left := visitExprList(node.Rhs, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderAssignStmt(at=%s), %s, %s", locationString, left, right)
	} else {
		return fmt.Sprintf("placeholderAssignStmt(), %s, %s", left, right)
	}
}

func visitGoStmt(node *ast.GoStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderGoStmt(at=%s), %s", locationString, visitCallExpr(node.Call, fset, addLocs))
	} else {
		return fmt.Sprintf("placeholderGoStmt(), %s", visitCallExpr(node.Call, fset, addLocs))
	}
}

func visitDeferStmt(node *ast.DeferStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderDeferStmt(at=%s), %s", locationString, visitCallExpr(node.Call, fset, addLocs))
	} else {
		return fmt.Sprintf("placeholderDeferStmt(), %s", visitCallExpr(node.Call, fset, addLocs))
	}
}

func visitReturnStmt(node *ast.ReturnStmt, fset *token.FileSet, addLocs bool) string {
	result := visitExprList(node.Results, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderReturnStmt(at=%s), %s", locationString, result)
	} else {
		return fmt.Sprintf("placeholderReturnStmt(), %s", result)
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
	stmts := visitFunctionBody(node, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderBlockStmt(at=%s), %s", locationString, stmts)
	} else {
		return fmt.Sprintf("placeholderBlockStmt(), %s", stmts)
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

func visitExpr(node *ast.Expr, fset *token.FileSet, addLocs bool) string {
	switch t := (*node).(type) {

	case *ast.Ident:
		return visitIdent(t, fset, addLocs)

	case *ast.Ellipsis:
		return visitEllipsis(t, fset, addLocs)

	case *ast.BasicLit:
		return visitBasicLit(t, fset, addLocs)

	case *ast.FuncLit:
		return visitFuncLit(t, fset, addLocs)

	case *ast.CompositeLit:
		return visitCompositeLit(t, fset, addLocs)

	case *ast.ParenExpr:
		return visitParenExpr(t, fset, addLocs)

	case *ast.SelectorExpr:
		return visitSelectorExpr(t, fset, addLocs)

	case *ast.IndexExpr:
		return visitIndexExpr(t, fset, addLocs)

	case *ast.IndexListExpr:
		return visitIndexListExpr(t, fset, addLocs)

	case *ast.SliceExpr:
		return visitSliceExpr(t, fset, addLocs)

	case *ast.TypeAssertExpr:
		return visitTypeAssertExpr(t, fset, addLocs)

	case *ast.CallExpr:
		return visitCallExpr(t, fset, addLocs)

	case *ast.StarExpr:
		return visitStarExpr(t, fset, addLocs)

	case *ast.UnaryExpr:
		return visitUnaryExpr(t, fset, addLocs)

	case *ast.BinaryExpr:
		return visitBinaryExpr(t, fset, addLocs)

	case *ast.KeyValueExpr:
		return visitKeyValueExpr(t, fset, addLocs)

	case *ast.ArrayType:
		return visitArrayType(t, fset, addLocs)

	case *ast.StructType:
		return visitStructType(t, fset, addLocs)

	case *ast.FuncType:
		return visitFuncType(t, fset, addLocs)

	case *ast.InterfaceType:
		return visitInterfaceType(t, fset, addLocs)

	case *ast.MapType:
		return visitMapType(t, fset, addLocs)

	case *ast.ChanType:
		return visitChanType(t, fset, addLocs)
	}
	return ""
}

func visitIdent(node *ast.Ident, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderIdent(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderIdent()")
	}
}
func visitEllipsis(node *ast.Ellipsis, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderEllipsis(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderEllipsis()")
	}
}
func visitBasicLit(node *ast.BasicLit, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderBasicLit(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderBasicLit()")
	}
}
func visitFuncLit(node *ast.FuncLit, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderFuncLit(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderFuncLit()")
	}
}
func visitCompositeLit(node *ast.CompositeLit, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderCompositeLit(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderCompositeLit()")
	}
}
func visitParenExpr(node *ast.ParenExpr, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderParenExpr(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderParenExpr()")
	}
}
func visitSelectorExpr(node *ast.SelectorExpr, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderSelectorExpr(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderSelectorExpr()")
	}
}
func visitIndexExpr(node *ast.IndexExpr, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderIndexExpr(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderIndexExpr()")
	}
}
func visitIndexListExpr(node *ast.IndexListExpr, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderIndexListExpr(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderIndexListExpr()")
	}
}
func visitSliceExpr(node *ast.SliceExpr, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderSliceExpr(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderSliceExpr()")
	}
}
func visitTypeAssertExpr(node *ast.TypeAssertExpr, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderTypeAssertExpr(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderTypeAssertExpr()")
	}
}
func visitCallExpr(node *ast.CallExpr, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderCallExpr(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderCallExpr()")
	}
}
func visitStarExpr(node *ast.StarExpr, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderStarExpr(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderStarExpr()")
	}
}
func visitUnaryExpr(node *ast.UnaryExpr, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderUnaryExpr(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderUnaryExpr()")
	}
}
func visitBinaryExpr(node *ast.BinaryExpr, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderBinaryExpr(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderBinaryExpr()")
	}
}
func visitKeyValueExpr(node *ast.KeyValueExpr, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderKeyValueExpr(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderKeyValueExpr()")
	}
}
func visitArrayType(node *ast.ArrayType, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderArrayType(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderArrayType()")
	}
}
func visitStructType(node *ast.StructType, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderStructType(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderStructType()")
	}
}
func visitFuncType(node *ast.FuncType, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderFuncType(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderFuncType()")
	}
}
func visitInterfaceType(node *ast.InterfaceType, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderInterfaceType(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderInterfaceType()")
	}
}
func visitMapType(node *ast.MapType, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderMapType(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderMapType()")
	}
}
func visitChanType(node *ast.ChanType, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("placeholderChanType(at=%s)", locationString)
	} else {
		return fmt.Sprintf("placeholderChanType()")
	}
}

func visitExprList(nodes []ast.Expr, fset *token.FileSet, addLocs bool) string {
	var exprs []string
	for i := 0; i < len(nodes); i++ {
		exprs = append(exprs, visitExpr(&nodes[i], fset, addLocs))
	}
	return strings.Join(exprs, ",")
}

// end of exprs
// ---------------------------------------------------------------------------------------------------------------
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
