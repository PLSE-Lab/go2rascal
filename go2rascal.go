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
		return fmt.Sprintf("genDecl(\"%s\", at=%s), ", node.Tok, locationString)
	} else {
		return fmt.Sprintf("genDecl(\"%s\"), ", node.Tok)
	}
}

func visitFunctionDeclaration(node *ast.FuncDecl, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("func(\"%s\", at=%s), [%s]", node.Name, locationString, visitFunctionBody(node.Body, fset, addLocs))
	} else {
		return fmt.Sprintf("func(\"%s\"), [%s]", node.Name, visitFunctionBody(node.Body, fset, addLocs))
	}
}

//end of declarations.
//---------------------------------------------------------------------------------------------------------------------------

func visitImportSpec(node *ast.ImportSpec, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("import(\"%s\", at=%s), ", node.Name, locationString)
	} else {
		return fmt.Sprintf("import(\"%s\"), ", node.Name)
	}
}

// end of imports.
// ----------------------------------------------------------------------------------------------------------------------------

func visitFunctionBody(node *ast.BlockStmt, fset *token.FileSet, addLocs bool) string {
	var stmts []string
	for i := 0; i < len(node.List); i++ {
		stmts = append(stmts, visitStmt(&node.List[i], fset, addLocs))
	}

	stmtsString := strings.Join(stmts, ",")
	if addLocs {
		return fmt.Sprintf("%s", stmtsString)
	} else {
		return fmt.Sprintf("%s", stmtsString)
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
		return fmt.Sprintf("%s, ", visitDeclaration(&node.Decl, fset, addLocs))
	} else {
		return fmt.Sprintf("%s, ", visitDeclaration(&node.Decl, fset, addLocs))
	}
}

// need to come back to I think
func visitEmptyStmt(node *ast.EmptyStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("emptyStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("emptyStmt()")
	}
}

func visitLabeledStmt(node *ast.LabeledStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("labeledStmt(\"%s\", at=%s), %s", node.Label, locationString, visitStmt(&node.Stmt, fset, addLocs))
	} else {
		return fmt.Sprintf("labeledStmt(\"%s\"), %s", node.Label, visitStmt(&node.Stmt, fset, addLocs))
	}
}

//-----------------------------------------------------------------

func visitExprStmt(node *ast.ExprStmt, fset *token.FileSet, addLocs bool) string {
	e := visitExpr(&node.X, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("exprStmt(%s, at=%s) ", e, locationString)
	} else {
		return fmt.Sprintf("exprStmt(%s) ", e)
	}
}

func visitSendStmt(node *ast.SendStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("sendStmt(%s, %s, at=%s) ", visitExpr(&node.Chan, fset, addLocs), visitExpr(&node.Value, fset, addLocs), locationString)
	} else {
		return fmt.Sprintf("sendStmt(%s, %s)", visitExpr(&node.Chan, fset, addLocs), visitExpr(&node.Value, fset, addLocs))
	}
}

func visitIncDecStmt(node *ast.IncDecStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("incDecStmt(\"%s\", %s, at=%s)", node.Tok, visitExpr(&node.X, fset, addLocs), locationString)
	} else {
		return fmt.Sprintf("incDecStmt(\"%s\", %s) ", node.Tok, visitExpr(&node.X, fset, addLocs))
	}
}

func visitAssignStmt(node *ast.AssignStmt, fset *token.FileSet, addLocs bool) string {
	right := visitExprList(node.Rhs, fset, addLocs)
	left := visitExprList(node.Rhs, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("assignStmt( %s, %s, at=%s) ", left, right, locationString)
	} else {
		return fmt.Sprintf("assignStmt(%s, %s) ", left, right)
	}
}

func visitGoStmt(node *ast.GoStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("goStmt(%s, at=%s) ", visitCallExpr(node.Call, fset, addLocs), locationString)
	} else {
		return fmt.Sprintf("goStmt(%s) ", visitCallExpr(node.Call, fset, addLocs))
	}
}

func visitDeferStmt(node *ast.DeferStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("deferStmt(%s, at=%s) ", visitCallExpr(node.Call, fset, addLocs), locationString)
	} else {
		return fmt.Sprintf("deferStmt(%s) ", visitCallExpr(node.Call, fset, addLocs))
	}
}

func visitReturnStmt(node *ast.ReturnStmt, fset *token.FileSet, addLocs bool) string {
	result := visitExprList(node.Results, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("returnStmt(%s, at=%s) ", result, locationString)
	} else {
		return fmt.Sprintf("returnStmt(%s) ", result)
	}
}

func visitBranchStmt(node *ast.BranchStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("branchStmt(%s, %s, at=%s) ", node.Tok, node.Label, locationString)
	} else {
		return fmt.Sprintf("branchStmt(%s, %s)", node.Tok, node.Label)
	}
}

func visitBlockStmt(node *ast.BlockStmt, fset *token.FileSet, addLocs bool) string {
	stmts := visitFunctionBody(node, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("blockStmt(%s, at=%s) ", stmts, locationString)
	} else {
		return fmt.Sprintf("blockStmt(%s)", stmts)
	}
}

func visitIfStmt(node *ast.IfStmt, fset *token.FileSet, addLocs bool) string {
	block := visitBlockStmt(node.Body, fset, addLocs)
	expr := visitExpr(&node.Cond, fset, addLocs)
	ifStmt := visitStmt(&node.Init, fset, addLocs)
	elseStmt := visitStmt(&node.Else, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("ifStmt(%s, %s, [%s], %s, at=%s) ", ifStmt, expr, block, elseStmt, locationString)
	} else {
		return fmt.Sprintf("ifStmt(%s, %s, [%s], %s) ", ifStmt, expr, block, elseStmt)
	}
}

func visitCaseClause(node *ast.CaseClause, fset *token.FileSet, addLocs bool) string {
	var stmts []string
	for i := 0; i < len(node.List); i++ {
		stmts = append(stmts, visitStmt(&node.Body[i], fset, addLocs))
	}
	exprs := visitExprList(node.List, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("caseClause([%s], [%s], at=%s) ", stmts, exprs, locationString)
	} else {
		return fmt.Sprintf("caseClause(%s, [%s]) ", stmts, exprs)
	}
}

func visitSwitchStmt(node *ast.SwitchStmt, fset *token.FileSet, addLocs bool) string {
	block := visitBlockStmt(node.Body, fset, addLocs)
	init := visitStmt(&node.Init, fset, addLocs)
	tag := visitExpr(&node.Tag, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("switchStmt(%s, %s, [%s], at=%s) ", init, tag, block, locationString)
	} else {
		return fmt.Sprintf("switchStmt(%s, %s, [%s]) ", init, tag, block)
	}
}

func visitTypeSwitchStmt(node *ast.TypeSwitchStmt, fset *token.FileSet, addLocs bool) string {
	block := visitBlockStmt(node.Body, fset, addLocs)
	init := visitStmt(&node.Init, fset, addLocs)
	assign := visitStmt(&node.Assign, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("typeSwitchStmt(%s, %s, [%s] at=%s) ", init, assign, block, locationString)
	} else {
		return fmt.Sprintf("typeSwitchStmt(%s, %s, [%s])", init, assign, block)
	}
}

func visitCommClause(node *ast.CommClause, fset *token.FileSet, addLocs bool) string {
	comm := visitStmt(&node.Comm, fset, addLocs)
	var stmts []string
	for i := 0; i < len(node.Body); i++ {
		stmts = append(stmts, visitStmt(&node.Body[i], fset, addLocs))
	}
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("commClause(%s, [%s], at=%s)", comm, stmts, locationString)
	} else {
		return fmt.Sprintf("commClause(%s, [%s])", comm, stmts)
	}
}

func visitSelectStmt(node *ast.SelectStmt, fset *token.FileSet, addLocs bool) string {
	block := visitBlockStmt(node.Body, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("selectStmt([%s], at=%s) ", block, locationString)
	} else {
		return fmt.Sprintf("selectStmt([%s]) ", block)
	}
}

func visitForStmt(node *ast.ForStmt, fset *token.FileSet, addLocs bool) string {
	block := visitBlockStmt(node.Body, fset, addLocs)
	init := visitStmt(&node.Init, fset, addLocs)
	post := visitStmt(&node.Post, fset, addLocs)
	cond := visitExpr(&node.Cond, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("forStmt(%s, %s, %s, [%s], at=%s)", init, cond, post, block, locationString)
	} else {
		return fmt.Sprintf("forStmt(%s, %s, %s, [%s])", init, cond, post, block)
	}
}

func visitRangeStmt(node *ast.RangeStmt, fset *token.FileSet, addLocs bool) string {
	key := visitExpr(&node.Key, fset, addLocs)
	value := visitExpr(&node.Value, fset, addLocs)
	x := visitExpr(&node.X, fset, addLocs)
	block := visitBlockStmt(node.Body, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("rangeStmt(%s, %s, %s, [%s], at=%s)", key, value, x, block, locationString)
	} else {
		return fmt.Sprintf("rangeStmt(%s, %s, %s, [%s])", key, value, x, block)
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
	name := node.Name
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("ident(%s, at=%s)", name, locationString)
	} else {
		return fmt.Sprintf("ident(%s)", name)
	}
}
func visitEllipsis(node *ast.Ellipsis, fset *token.FileSet, addLocs bool) string {
	elt := visitExpr(&node.Elt, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("ellipsis(%s, at=%s)", elt, locationString)
	} else {
		return fmt.Sprintf("ellipsis(%s)", elt)
	}
}
func visitBasicLit(node *ast.BasicLit, fset *token.FileSet, addLocs bool) string {
	value := node.Value
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("basicLit(%s, at=%s)", value, locationString)
	} else {
		return fmt.Sprintf("basicLit(%s)", value)
	}
}
func visitFuncLit(node *ast.FuncLit, fset *token.FileSet, addLocs bool) string {
	body := visitBlockStmt(node.Body, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("funcLit([%s], at=%s)", body, locationString)
	} else {
		return fmt.Sprintf("funcLit(%s)", body)
	}
}
func visitCompositeLit(node *ast.CompositeLit, fset *token.FileSet, addLocs bool) string {
	types := visitExpr(&node.Type, fset, addLocs)
	elts := visitExprList(node.Elts, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("compositeLit(%s, [%s], at=%s)", types, elts, locationString)
	} else {
		return fmt.Sprintf("compositeLit(%s, [%s])", types, elts)
	}
}
func visitParenExpr(node *ast.ParenExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("parenExpr(%s, at=%s)", x, locationString)
	} else {
		return fmt.Sprintf("parenExpr(%s)", x)
	}
}
func visitSelectorExpr(node *ast.SelectorExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("selectorExpr(%s, at=%s)", x, locationString)
	} else {
		return fmt.Sprintf("selectorExpr(%s)", x)
	}
}
func visitIndexExpr(node *ast.IndexExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	index := visitExpr(&node.Index, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("indexExpr(%s, %s, at=%s)", x, index, locationString)
	} else {
		return fmt.Sprintf("indexExpr(%s, %s)", x, index)
	}
}
func visitIndexListExpr(node *ast.IndexListExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	indices := visitExprList(node.Indices, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("indexListExpr(%s, [%s], at=%s)", x, indices, locationString)
	} else {
		return fmt.Sprintf("indexListExpr(%s, [%s])", x, indices)
	}
}
func visitSliceExpr(node *ast.SliceExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	low := visitExpr(&node.Low, fset, addLocs)
	high := visitExpr(&node.High, fset, addLocs)
	max := visitExpr(&node.Max, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("sliceExpr(%s, %s, %s, %s, at=%s)", x, low, high, max, locationString)
	} else {
		return fmt.Sprintf("sliceExpr(%s, %s, %s, %s)", x, low, high, max)
	}
}
func visitTypeAssertExpr(node *ast.TypeAssertExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	types := visitExpr(&node.Type, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("typeAssertExpr(%s, %s, at=%s)", x, types, locationString)
	} else {
		return fmt.Sprintf("typeAssertExpr(%s, %s)", x, types)
	}
}
func visitCallExpr(node *ast.CallExpr, fset *token.FileSet, addLocs bool) string {
	fun := visitExpr(&node.Fun, fset, addLocs)
	args := visitExprList(node.Args, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("callExpr(%s, [%s], at=%s)", fun, args, locationString)
	} else {
		return fmt.Sprintf("callExpr(%s, [%s])", fun, args)
	}
}
func visitStarExpr(node *ast.StarExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("starExpr(%s, at=%s)", x, locationString)
	} else {
		return fmt.Sprintf("starExpr(%s)", x)
	}
}
func visitUnaryExpr(node *ast.UnaryExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	tok := node.Op.String()
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("unaryExpr(%s, %s, at=%s)", x, tok, locationString)
	} else {
		return fmt.Sprintf("unaryExpr(%s, %s)", x, tok)
	}
}
func visitBinaryExpr(node *ast.BinaryExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	tok := node.Op.String()
	y := visitExpr(&node.Y, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("binaryExpr(%s, %s, %s, at=%s)", x, tok, y, locationString)
	} else {
		return fmt.Sprintf("binaryExpr(%s, %s, %s)", x, tok, y)
	}
}
func visitKeyValueExpr(node *ast.KeyValueExpr, fset *token.FileSet, addLocs bool) string {
	key := visitExpr(&node.Key, fset, addLocs)
	value := visitExpr(&node.Value, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("keyValueExpr(%s, %s, at=%s)", key, value, locationString)
	} else {
		return fmt.Sprintf("keyValueExpr(%s, %s)", key, value)
	}
}
func visitArrayType(node *ast.ArrayType, fset *token.FileSet, addLocs bool) string {
	lens := visitExpr(&node.Len, fset, addLocs)
	elt := visitExpr(&node.Elt, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("arrayType(%s, %s, at=%s)", lens, elt, locationString)
	} else {
		return fmt.Sprintf("arrayType(%s, %s)", lens, elt)
	}
}
func visitStructType(node *ast.StructType, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("structType(at=%s)", locationString)
	} else {
		return fmt.Sprintf("structType()")
	}
}
func visitFuncType(node *ast.FuncType, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("funcType(at=%s)", locationString)
	} else {
		return fmt.Sprintf("funcType()")
	}
}
func visitInterfaceType(node *ast.InterfaceType, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("interfaceType(at=%s)", locationString)
	} else {
		return fmt.Sprintf("interfaceType()")
	}
}
func visitMapType(node *ast.MapType, fset *token.FileSet, addLocs bool) string {
	key := visitExpr(&node.Key, fset, addLocs)
	value := visitExpr(&node.Value, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("mapType(%s, %s, at=%s)", key, value, locationString)
	} else {
		return fmt.Sprintf("mapType(%s, %s)", key, value)
	}
}
func visitChanType(node *ast.ChanType, fset *token.FileSet, addLocs bool) string {
	value := visitExpr(&node.Value, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("chanType(%s, at=%s)", value, locationString)
	} else {
		return fmt.Sprintf("chanType()")
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
