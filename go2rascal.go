package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

var rascalizer *strings.Replacer = strings.NewReplacer("<", "\\<", ">", "\\>", "\n", "\\n", "\t", "\\t", "\r", "\\r", "\\", "\\\\", "\"", "\\\"", "'", "\\'")

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

func visitFile(node *ast.File, fset *token.FileSet, addLocs bool) string {
	var decls []string
	for i := 0; i < len(node.Decls); i++ {
		decls = append(decls, visitDeclaration(&node.Decls[i], fset, addLocs))
	}
	declString := strings.Join(decls, ",")

	// These are already part of the declarations...
	//var imports []string
	//for i := 0; i < len(node.Imports); i++ {
	//	imports = append(imports, visitImportSpec(node.Imports[i], fset, addLocs))
	//}
	//importString := strings.Join(imports, ",")

	packageName := node.Name.Name

	if addLocs {
		locationString := computeLocation(fset, node.FileStart, node.FileEnd)
		return fmt.Sprintf("file(\"%s\", [%s], at=%s)", packageName, declString, locationString)
	} else {
		return fmt.Sprintf("file(\"%s\", [%s])", packageName, declString)
	}

}

func visitDeclaration(node *ast.Decl, fset *token.FileSet, addLocs bool) string {
	switch d := (*node).(type) {
	case *ast.GenDecl:
		return visitGeneralDeclaration(d, fset, addLocs)
	case *ast.FuncDecl:
		return visitFunctionDeclaration(d, fset, addLocs)
	default:
		return "unknownDeclaration()" // This is an error, should panic here
	}
}

func optionalNameToRascal(node *ast.Ident) string {
	if node != nil {
		return fmt.Sprintf("someName(\"%s\")", node.Name)
	} else {
		return "noName()"
	}
}

func visitSpec(node *ast.Spec, fset *token.FileSet, addLocs bool) string {
	switch d := (*node).(type) {
	case *ast.ImportSpec:
		return visitImportSpec(d, fset, addLocs)
	case *ast.ValueSpec:
		return visitValueSpec(d, fset, addLocs)
	case *ast.TypeSpec:
		return visitTypeSpec(d, fset, addLocs)
	default:
		return "unknownSpec()" // This is an error, should panic here
	}
}

func visitImportSpec(node *ast.ImportSpec, fset *token.FileSet, addLocs bool) string {
	specName := optionalNameToRascal(node.Name)
	specPath := literalToRascal(node.Path, fset, addLocs)

	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("importSpec(%s,%s,at=%s)", specName, specPath, locationString)
	} else {
		return fmt.Sprintf("importSpec(%s,%s)", specName, specPath)
	}
}

func visitValueSpec(node *ast.ValueSpec, fset *token.FileSet, addLocs bool) string {
	var names []string
	for i := 0; i < len(node.Names); i++ {
		names = append(names, fmt.Sprintf("\"%s\"", node.Names[i].Name))
	}
	namesStr := fmt.Sprintf("[%s]", strings.Join(names, ","))
	typeStr := visitOptionExpr(&node.Type, fset, addLocs)
	values := visitExprList(node.Values, fset, addLocs)

	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("valueSpec(%s,%s,%s,at=%s)", namesStr, typeStr, values, locationString)
	} else {
		return fmt.Sprintf("valueSpec(%s,%s,%s)", namesStr, typeStr, values)
	}
}

func visitTypeSpec(node *ast.TypeSpec, fset *token.FileSet, addLocs bool) string {
	typeParams := visitFieldList(node.TypeParams, fset, addLocs)
	typeStr := visitExpr(&node.Type, fset, addLocs)

	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("typeSpec(\"%s\",%s,%s,at=%s)", node.Name.Name, typeParams, typeStr, locationString)
	} else {
		return fmt.Sprintf("typeSpec(\"%s\",%s,%s)", node.Name.Name, typeParams, typeStr)
	}
}

func visitSpecList(nodes []ast.Spec, fset *token.FileSet, addLocs bool) string {
	var specs []string
	for i := 0; nodes != nil && i < len(nodes); i++ {
		specs = append(specs, visitSpec(&nodes[i], fset, addLocs))
	}
	return fmt.Sprintf("[%s]", strings.Join(specs, ","))
}

func visitGeneralDeclaration(node *ast.GenDecl, fset *token.FileSet, addLocs bool) string {
	declType := declTypeToRascal(node.Tok)
	specList := visitSpecList(node.Specs, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("genDecl(%s,%s,at=%s)", declType, specList, locationString)
	} else {
		return fmt.Sprintf("genDecl(%s,%s)", declType, specList)
	}
}

func visitFunctionDeclaration(node *ast.FuncDecl, fset *token.FileSet, addLocs bool) string {
	receivers := visitFieldList(node.Recv, fset, addLocs)
	signature := visitFuncType(node.Type, fset, addLocs)
	body := visitOptionalBlockStmt(node.Body, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("funDecl(\"%s\",%s,%s,%s,at=%s)", node.Name, receivers, signature, body, locationString)
	} else {
		return fmt.Sprintf("funDecl(\"%s\",%s,%s,%s)", node.Name, receivers, signature, body)
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
	default:
		return fmt.Sprintf("unknownStmt(\"%s\")", reflect.TypeOf(node).Name())
	}
}

func visitOptionStmt(node *ast.Stmt, fset *token.FileSet, addLocs bool) string {
	if *node == nil {
		return "noStmt()"
	} else {
		return fmt.Sprintf("someStmt(%s)", visitStmt(node, fset, addLocs))
	}
}

func visitOptionExpr(node *ast.Expr, fset *token.FileSet, addLocs bool) string {
	if *node == nil {
		return "noExpr()"
	} else {
		return fmt.Sprintf("someExpr(%s)", visitExpr(node, fset, addLocs))
	}
}

func visitDeclStmt(node *ast.DeclStmt, fset *token.FileSet, addLocs bool) string {
	declStr := visitDeclaration(&node.Decl, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("declStmt(%s,at=%s)", declStr, locationString)
	} else {
		return fmt.Sprintf("declStmt(%s)", declStr)
	}
}

func visitEmptyStmt(node *ast.EmptyStmt, fset *token.FileSet, addLocs bool) string {
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("emptyStmt(at=%s)", locationString)
	} else {
		return fmt.Sprintf("emptyStmt()")
	}
}

func visitLabeledStmt(node *ast.LabeledStmt, fset *token.FileSet, addLocs bool) string {
	labelStr := labelToRascal(node.Label)
	stmtStr := visitStmt(&node.Stmt, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("labeledStmt(%s,%s,at=%s)", labelStr, stmtStr, locationString)
	} else {
		return fmt.Sprintf("labeledStmt(%s,%s)", labelStr, stmtStr)
	}
}

func visitExprStmt(node *ast.ExprStmt, fset *token.FileSet, addLocs bool) string {
	exprStr := visitExpr(&node.X, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("exprStmt(%s,at=%s)", exprStr, locationString)
	} else {
		return fmt.Sprintf("exprStmt(%s)", exprStr)
	}
}

func visitSendStmt(node *ast.SendStmt, fset *token.FileSet, addLocs bool) string {
	chanStr := visitExpr(&node.Chan, fset, addLocs)
	valStr := visitExpr(&node.Value, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("sendStmt(%s,%s,at=%s)", chanStr, valStr, locationString)
	} else {
		return fmt.Sprintf("sendStmt(%s,%s)", chanStr, valStr)
	}
}

func visitIncDecStmt(node *ast.IncDecStmt, fset *token.FileSet, addLocs bool) string {
	opStr := opToRascal(node.Tok)
	exprStr := visitExpr(&node.X, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("incDecStmt(%s,%s,at=%s)", opStr, exprStr, locationString)
	} else {
		return fmt.Sprintf("incDecStmt(%s,%s)", opStr, exprStr)
	}
}

func visitAssignStmt(node *ast.AssignStmt, fset *token.FileSet, addLocs bool) string {
	assignOp := assignmentOpToRascal(node.Tok)
	left := visitExprList(node.Lhs, fset, addLocs)
	right := visitExprList(node.Rhs, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("assignStmt(%s,%s,%s,at=%s)", left, right, assignOp, locationString)
	} else {
		return fmt.Sprintf("assignStmt(%s,%s,%s)", left, right, assignOp)
	}
}

func visitGoStmt(node *ast.GoStmt, fset *token.FileSet, addLocs bool) string {
	exprStr := visitCallExpr(node.Call, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("goStmt(%s,at=%s)", exprStr, locationString)
	} else {
		return fmt.Sprintf("goStmt(%s)", exprStr)
	}
}

func visitDeferStmt(node *ast.DeferStmt, fset *token.FileSet, addLocs bool) string {
	exprStr := visitCallExpr(node.Call, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("deferStmt(%s,at=%s)", exprStr, locationString)
	} else {
		return fmt.Sprintf("deferStmt(%s)", exprStr)
	}
}

func visitReturnStmt(node *ast.ReturnStmt, fset *token.FileSet, addLocs bool) string {
	results := visitExprList(node.Results, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("returnStmt(%s,at=%s)", results, locationString)
	} else {
		return fmt.Sprintf("returnStmt(%s)", results)
	}
}

func visitBranchStmt(node *ast.BranchStmt, fset *token.FileSet, addLocs bool) string {
	typeStr := branchTypeToRascal(node.Tok)
	labelStr := labelToRascal(node.Label)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("branchStmt(%s,%s,at=%s)", typeStr, labelStr, locationString)
	} else {
		return fmt.Sprintf("branchStmt(%s,%s)", typeStr, labelStr)
	}
}

func visitOptionalBlockStmt(node *ast.BlockStmt, fset *token.FileSet, addLocs bool) string {
	if node == nil {
		return "noStmt()"
	} else {
		return fmt.Sprintf("someStmt(%s)", visitBlockStmt(node, fset, addLocs))
	}
}

func visitBlockStmt(node *ast.BlockStmt, fset *token.FileSet, addLocs bool) string {
	stmts := visitStmtList(node.List, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("blockStmt(%s,at=%s)", stmts, locationString)
	} else {
		return fmt.Sprintf("blockStmt(%s)", stmts)
	}
}

func visitIfStmt(node *ast.IfStmt, fset *token.FileSet, addLocs bool) string {
	initStmt := visitOptionStmt(&node.Init, fset, addLocs)
	condExpr := visitExpr(&node.Cond, fset, addLocs)
	body := visitBlockStmt(node.Body, fset, addLocs)
	elseStmt := visitOptionStmt(&node.Else, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("ifStmt(%s,%s,%s,%s,at=%s)", initStmt, condExpr, body, elseStmt, locationString)
	} else {
		return fmt.Sprintf("ifStmt(%s,%s,%s,%s)", initStmt, condExpr, body, elseStmt)
	}
}

func caseToRascal(nodes []ast.Expr, fset *token.FileSet, addLocs bool) string {
	if nodes != nil {
		return fmt.Sprintf("regularCase(%s)", visitExprList(nodes, fset, addLocs))
	} else {
		return "defaultCase()"
	}
}

func visitCaseClause(node *ast.CaseClause, fset *token.FileSet, addLocs bool) string {
	caseStr := caseToRascal(node.List, fset, addLocs)
	stmtsString := visitStmtList(node.Body, fset, addLocs)

	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("caseClause(%s,%s,at=%s)", caseStr, stmtsString, locationString)
	} else {
		return fmt.Sprintf("caseClause(%s,%s)", caseStr, stmtsString)
	}
}

func visitCaseClauseList(node *ast.BlockStmt, fset *token.FileSet, addLocs bool) string {
	var cases []string
	for i := 0; node != nil && node.List != nil && i < len(node.List); i++ {
		nodeAsCase, ok := node.List[i].(*ast.CaseClause)
		if ok {
			cases = append(cases, visitCaseClause(nodeAsCase, fset, addLocs))
		} else {
			cases = append(cases, "invalidCaseClause()")
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(cases, ","))
}

func visitSwitchStmt(node *ast.SwitchStmt, fset *token.FileSet, addLocs bool) string {
	init := visitOptionStmt(&node.Init, fset, addLocs)
	tag := visitOptionExpr(&node.Tag, fset, addLocs)
	block := visitCaseClauseList(node.Body, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("switchStmt(%s,%s,%s,at=%s)", init, tag, block, locationString)
	} else {
		return fmt.Sprintf("switchStmt(%s,%s,%s)", init, tag, block)
	}
}

func visitTypeSwitchStmt(node *ast.TypeSwitchStmt, fset *token.FileSet, addLocs bool) string {
	init := visitOptionStmt(&node.Init, fset, addLocs)
	assign := visitStmt(&node.Assign, fset, addLocs)
	block := visitCaseClauseList(node.Body, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("typeSwitchStmt(%s,%s,%s,at=%s)", init, assign, block, locationString)
	} else {
		return fmt.Sprintf("typeSwitchStmt(%s,%s,%s)", init, assign, block)
	}
}

func clauseToRascal(node *ast.Stmt, fset *token.FileSet, addLocs bool) string {
	if node == nil {
		return "defaultComm()"
	} else {
		return fmt.Sprintf("regularComm(%s)", visitStmt(node, fset, addLocs))
	}
}

func visitCommClause(node *ast.CommClause, fset *token.FileSet, addLocs bool) string {
	clauseStr := clauseToRascal(&node.Comm, fset, addLocs)
	stmtsString := visitStmtList(node.Body, fset, addLocs)

	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("commClause(%s,%s,at=%s)", clauseStr, stmtsString, locationString)
	} else {
		return fmt.Sprintf("commClause(%s,%s)", clauseStr, stmtsString)
	}
}

func visitCommClauseList(node *ast.BlockStmt, fset *token.FileSet, addLocs bool) string {
	var clauses []string
	for i := 0; node != nil && node.List != nil && i < len(node.List); i++ {
		nodeAsClause, ok := node.List[i].(*ast.CommClause)
		if ok {
			clauses = append(clauses, visitCommClause(nodeAsClause, fset, addLocs))
		} else {
			clauses = append(clauses, "invalidCommClause()")
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(clauses, ","))
}

func visitSelectStmt(node *ast.SelectStmt, fset *token.FileSet, addLocs bool) string {
	clauses := visitCommClauseList(node.Body, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("selectStmt(%s,at=%s)", clauses, locationString)
	} else {
		return fmt.Sprintf("selectStmt(%s)", clauses)
	}
}

func visitForStmt(node *ast.ForStmt, fset *token.FileSet, addLocs bool) string {
	init := visitOptionStmt(&node.Init, fset, addLocs)
	post := visitOptionStmt(&node.Post, fset, addLocs)
	cond := visitOptionExpr(&node.Cond, fset, addLocs)
	block := visitBlockStmt(node.Body, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("forStmt(%s,%s,%s,%s,at=%s)", init, cond, post, block, locationString)
	} else {
		return fmt.Sprintf("forStmt(%s,%s,%s,%s)", init, cond, post, block)
	}
}

func visitRangeStmt(node *ast.RangeStmt, fset *token.FileSet, addLocs bool) string {
	key := visitOptionExpr(&node.Key, fset, addLocs)
	value := visitOptionExpr(&node.Value, fset, addLocs)
	assignOp := assignmentOpToRascal(node.Tok)
	rangeExpr := visitExpr(&node.X, fset, addLocs)
	block := visitBlockStmt(node.Body, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("rangeStmt(%s,%s,%s,%s,%s,at=%s)", key, value, assignOp, rangeExpr, block, locationString)
	} else {
		return fmt.Sprintf("rangeStmt(%s,%s,%s,%s,%s)", key, value, assignOp, rangeExpr, block)
	}
}

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

	default:
		return fmt.Sprintf("unknownExpr(\"%s\")", reflect.TypeOf(node).Name())
	}
}

func visitIdent(node *ast.Ident, fset *token.FileSet, addLocs bool) string {
	name := node.Name
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("ident(\"%s\",at=%s)", name, locationString)
	} else {
		return fmt.Sprintf("ident(\"%s\")", name)
	}
}
func visitEllipsis(node *ast.Ellipsis, fset *token.FileSet, addLocs bool) string {
	elt := visitOptionExpr(&node.Elt, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("ellipsis(%s,at=%s)", elt, locationString)
	} else {
		return fmt.Sprintf("ellipsis(%s)", elt)
	}
}

func literalToRascal(node *ast.BasicLit, fset *token.FileSet, addLocs bool) string {
	locationString := computeLocation(fset, node.Pos(), node.End())

	switch node.Kind {
	case token.INT:
		if parsed, err := strconv.ParseInt(node.Value, 0, 0); err == nil {
			if addLocs {
				return fmt.Sprintf("literalInt(%d,at=%s)", parsed, locationString)
			} else {
				return fmt.Sprintf("literalInt(%d)", parsed)
			}
		} else if parsed, err := strconv.ParseUint(node.Value, 0, 0); err == nil {
			if addLocs {
				return fmt.Sprintf("literalInt(%d,at=%s)", parsed, locationString)
			} else {
				return fmt.Sprintf("literalInt(%d)", parsed)
			}
		} else {
			if addLocs {
				return fmt.Sprintf("unknownLiteral(\"%s\",at=%s)", node.Value, locationString)
			} else {
				return fmt.Sprintf("unknownLiteral(\"%s\")", node.Value)
			}
		}
	case token.FLOAT:
		if parsed, err := strconv.ParseFloat(node.Value, 64); err == nil {
			if addLocs {
				return fmt.Sprintf("literalFloat(%f,at=%s)", parsed, locationString)
			} else {
				return fmt.Sprintf("literalFloat(%f)", parsed)
			}
		} else {
			if addLocs {
				return fmt.Sprintf("unknownLiteral(\"%s\",at=%s)", node.Value, locationString)
			} else {
				return fmt.Sprintf("unknownLiteral(\"%s\")", node.Value)
			}
		}
	case token.CHAR:
		if addLocs {
			return fmt.Sprintf("literalChar(%s,at=%s)", rascalizeChar(node.Value), locationString)
		} else {
			return fmt.Sprintf("literalChar(%s)", rascalizeChar(node.Value))
		}
	case token.STRING:
		if addLocs {
			return fmt.Sprintf("literalString(%s,at=%s)", rascalizeString(node.Value), locationString)
		} else {
			return fmt.Sprintf("literalString(%s)", rascalizeString(node.Value))
		}
	case token.IMAG:
		if ic, err := strconv.ParseComplex(node.Value, 128); err == nil {
			if addLocs {
				return fmt.Sprintf("literalImaginary(%f,%f,at=%s)", real(ic), imag(ic), locationString)
			} else {
				return fmt.Sprintf("literalImaginary(%f,%f)", real(ic), imag(ic))
			}
		} else {
			if addLocs {
				return fmt.Sprintf("unknownLiteral(\"%s\",at=%s)", node.Value, locationString)
			} else {
				return fmt.Sprintf("unknownLiteral(\"%s\")", node.Value)
			}
		}
	default:
		if addLocs {
			return fmt.Sprintf("unknownLiteral(\"%s\",at=%s)", node.Value, locationString)
		} else {
			return fmt.Sprintf("unknownLiteral(\"%s\")", node.Value)
		}
	}
}

func visitBasicLit(node *ast.BasicLit, fset *token.FileSet, addLocs bool) string {
	value := literalToRascal(node, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("basicLit(%s,at=%s)", value, locationString)
	} else {
		return fmt.Sprintf("basicLit(%s)", value)
	}
}

func visitFuncLit(node *ast.FuncLit, fset *token.FileSet, addLocs bool) string {
	typeStr := visitFuncType(node.Type, fset, addLocs)
	body := visitBlockStmt(node.Body, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("funcLit(%s,%s,at=%s)", typeStr, body, locationString)
	} else {
		return fmt.Sprintf("funcLit(%s,%s)", typeStr, body)
	}
}
func boolToRascal(val bool) string {
	if val {
		return "true"
	} else {
		return "false"
	}
}
func visitCompositeLit(node *ast.CompositeLit, fset *token.FileSet, addLocs bool) string {
	typeStr := visitOptionExpr(&node.Type, fset, addLocs)
	elts := visitExprList(node.Elts, fset, addLocs)

	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("compositeLit(%s,%s,%s,at=%s)", typeStr, elts, boolToRascal(node.Incomplete), locationString)
	} else {
		return fmt.Sprintf("compositeLit(%s,%s,%s)", typeStr, elts, boolToRascal(node.Incomplete))
	}
}

func visitParenExpr(node *ast.ParenExpr, fset *token.FileSet, addLocs bool) string {
	// We are building a tree, we do not need to keep explicit parens
	return visitExpr(&node.X, fset, addLocs)
}

func visitSelectorExpr(node *ast.SelectorExpr, fset *token.FileSet, addLocs bool) string {
	exprStr := visitExpr(&node.X, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("selectorExpr(%s,\"%s\",at=%s)", exprStr, node.Sel.Name, locationString)
	} else {
		return fmt.Sprintf("selectorExpr(%s,\"%s\")", exprStr, node.Sel.Name)
	}
}
func visitIndexExpr(node *ast.IndexExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	index := visitExpr(&node.Index, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("indexExpr(%s,%s,at=%s)", x, index, locationString)
	} else {
		return fmt.Sprintf("indexExpr(%s,%s)", x, index)
	}
}
func visitIndexListExpr(node *ast.IndexListExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	indices := visitExprList(node.Indices, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("indexListExpr(%s,%s,at=%s)", x, indices, locationString)
	} else {
		return fmt.Sprintf("indexListExpr(%s,%s)", x, indices)
	}
}
func visitSliceExpr(node *ast.SliceExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	low := visitOptionExpr(&node.Low, fset, addLocs)
	high := visitOptionExpr(&node.High, fset, addLocs)
	max := visitOptionExpr(&node.Max, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("sliceExpr(%s,%s,%s,%s,%s,at=%s)", x, low, high, max, boolToRascal(node.Slice3), locationString)
	} else {
		return fmt.Sprintf("sliceExpr(%s,%s,%s,%s,%s)", x, low, high, max, boolToRascal(node.Slice3))
	}
}
func visitTypeAssertExpr(node *ast.TypeAssertExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	types := visitExpr(&node.Type, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("typeAssertExpr(%s,%s,at=%s)", x, types, locationString)
	} else {
		return fmt.Sprintf("typeAssertExpr(%s,%s)", x, types)
	}
}
func visitCallExpr(node *ast.CallExpr, fset *token.FileSet, addLocs bool) string {
	fun := visitExpr(&node.Fun, fset, addLocs)
	args := visitExprList(node.Args, fset, addLocs)
	hasEllipses := boolToRascal(node.Ellipsis != token.NoPos)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("callExpr(%s,%s,%s,at=%s)", fun, args, hasEllipses, locationString)
	} else {
		return fmt.Sprintf("callExpr(%s,%s,%s)", fun, args, hasEllipses)
	}
}
func visitStarExpr(node *ast.StarExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("starExpr(%s,at=%s)", x, locationString)
	} else {
		return fmt.Sprintf("starExpr(%s)", x)
	}
}
func visitUnaryExpr(node *ast.UnaryExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	tok := opToRascal(node.Op)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("unaryExpr(%s,%s, at=%s)", x, tok, locationString)
	} else {
		return fmt.Sprintf("unaryExpr(%s,%s)", x, tok)
	}
}
func visitBinaryExpr(node *ast.BinaryExpr, fset *token.FileSet, addLocs bool) string {
	x := visitExpr(&node.X, fset, addLocs)
	tok := opToRascal(node.Op)
	y := visitExpr(&node.Y, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("binaryExpr(%s,%s,%s,at=%s)", x, y, tok, locationString)
	} else {
		return fmt.Sprintf("binaryExpr(%s,%s,%s)", x, y, tok)
	}
}
func visitKeyValueExpr(node *ast.KeyValueExpr, fset *token.FileSet, addLocs bool) string {
	key := visitExpr(&node.Key, fset, addLocs)
	value := visitExpr(&node.Value, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("keyValueExpr(%s,%s,at=%s)", key, value, locationString)
	} else {
		return fmt.Sprintf("keyValueExpr(%s,%s)", key, value)
	}
}
func visitArrayType(node *ast.ArrayType, fset *token.FileSet, addLocs bool) string {
	lens := visitOptionExpr(&node.Len, fset, addLocs)
	elt := visitExpr(&node.Elt, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("arrayType(%s,%s,at=%s)", lens, elt, locationString)
	} else {
		return fmt.Sprintf("arrayType(%s,%s)", lens, elt)
	}
}

func visitOptionBasicLiteral(literal *ast.BasicLit, fset *token.FileSet, addLocs bool) string {
	if literal == nil {
		return "noLiteral()"
	} else {
		return fmt.Sprintf("someLiteral(%s)", literalToRascal(literal, fset, addLocs))
	}
}

func fieldToRascal(field *ast.Field, fset *token.FileSet, addLocs bool) string {
	var names []string
	for i := 0; field.Names != nil && i < len(field.Names); i++ {
		names = append(names, fmt.Sprintf("\"%s\"", field.Names[i].Name))
	}
	namesStr := fmt.Sprintf("[%s]", strings.Join(names, ","))
	fieldType := visitOptionExpr(&field.Type, fset, addLocs)
	fieldTag := visitOptionBasicLiteral(field.Tag, fset, addLocs)

	if addLocs {
		locationString := computeLocation(fset, field.Pos(), field.End())
		return fmt.Sprintf("field(%s,%s,%s,at=%s)", namesStr, fieldType, fieldTag, locationString)
	} else {
		return fmt.Sprintf("field(%s,%s,%s)", namesStr, fieldType, fieldTag)
	}
}

func visitFieldList(fieldList *ast.FieldList, fset *token.FileSet, addLocs bool) string {
	var fields []string
	for i := 0; fieldList != nil && i < len(fieldList.List); i++ {
		fields = append(fields, fieldToRascal(fieldList.List[i], fset, addLocs))
	}
	return fmt.Sprintf("[%s]", strings.Join(fields, ","))
}

func visitStructType(node *ast.StructType, fset *token.FileSet, addLocs bool) string {
	fieldStr := visitFieldList(node.Fields, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("structType(%s,at=%s)", fieldStr, locationString)
	} else {
		return fmt.Sprintf("structType(%s)", fieldStr)
	}
}
func visitFuncType(node *ast.FuncType, fset *token.FileSet, addLocs bool) string {
	typeParams := visitFieldList(node.TypeParams, fset, addLocs)
	params := visitFieldList(node.Params, fset, addLocs)
	returns := visitFieldList(node.Results, fset, addLocs)

	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("funcType(%s,%s,%s,at=%s)", typeParams, params, returns, locationString)
	} else {
		return fmt.Sprintf("funcType(%s,%s,%s)", typeParams, params, returns)
	}
}
func visitInterfaceType(node *ast.InterfaceType, fset *token.FileSet, addLocs bool) string {
	methods := visitFieldList(node.Methods, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("interfaceType(%s,at=%s)", methods, locationString)
	} else {
		return fmt.Sprintf("interfaceType(%s)", methods)
	}
}
func visitMapType(node *ast.MapType, fset *token.FileSet, addLocs bool) string {
	key := visitExpr(&node.Key, fset, addLocs)
	value := visitExpr(&node.Value, fset, addLocs)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("mapType(%s,%s,at=%s)", key, value, locationString)
	} else {
		return fmt.Sprintf("mapType(%s,%s)", key, value)
	}
}
func visitChanType(node *ast.ChanType, fset *token.FileSet, addLocs bool) string {
	value := visitExpr(&node.Value, fset, addLocs)
	chanSend := boolToRascal(node.Dir == ast.SEND)
	if addLocs {
		locationString := computeLocation(fset, node.Pos(), node.End())
		return fmt.Sprintf("chanType(%s,%s,at=%s)", value, chanSend, locationString)
	} else {
		return fmt.Sprintf("chanType(%s,%s)", value, chanSend)
	}
}

func visitExprList(nodes []ast.Expr, fset *token.FileSet, addLocs bool) string {
	var exprs []string
	for i := 0; nodes != nil && i < len(nodes); i++ {
		exprs = append(exprs, visitExpr(&nodes[i], fset, addLocs))
	}
	return fmt.Sprintf("[%s]", strings.Join(exprs, ","))
}

func visitStmtList(nodes []ast.Stmt, fset *token.FileSet, addLocs bool) string {
	var stmts []string
	for i := 0; nodes != nil && i < len(nodes); i++ {
		stmts = append(stmts, visitStmt(&nodes[i], fset, addLocs))
	}

	return fmt.Sprintf("[%s]", strings.Join(stmts, ","))
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

func opToRascal(node token.Token) string {
	switch node {
	case token.ADD:
		return "add()"
	case token.SUB:
		return "sub()"
	case token.MUL:
		return "mul()"
	case token.QUO:
		return "quo()"
	case token.REM:
		return "rem()"
	case token.AND:
		return "and()"
	case token.OR:
		return "or()"
	case token.XOR:
		return "xor()"
	case token.SHL:
		return "shiftLeft()"
	case token.SHR:
		return "shiftRight()"
	case token.AND_NOT:
		return "andNot()"
	case token.LAND:
		return "logicalAnd()"
	case token.LOR:
		return "logicalOr()"
	case token.ARROW:
		return "arrow()"
	case token.INC:
		return "inc()"
	case token.DEC:
		return "dec()"
	case token.EQL:
		return "equal()"
	case token.LSS:
		return "lessThan()"
	case token.GTR:
		return "greaterThan()"
	case token.NOT:
		return "not()"
	case token.NEQ:
		return "notEqual()"
	case token.LEQ:
		return "lessThanEq()"
	case token.GEQ:
		return "greaterThanEq()"
	default:
		return fmt.Sprintf("unknownOp(%s)", node.String())
	}
}

func assignmentOpToRascal(node token.Token) string {
	switch node {
	case token.ADD_ASSIGN:
		return "addAssign()"
	case token.SUB_ASSIGN:
		return "subAssign()"
	case token.MUL_ASSIGN:
		return "mulAssign()"
	case token.QUO_ASSIGN:
		return "quoAssign()"
	case token.REM_ASSIGN:
		return "remAssign()"
	case token.AND_ASSIGN:
		return "andAssign()"
	case token.OR_ASSIGN:
		return "orAssign()"
	case token.XOR_ASSIGN:
		return "xorAssign()"
	case token.SHL_ASSIGN:
		return "shiftLeftAssign()"
	case token.SHR_ASSIGN:
		return "shiftRightAssign()"
	case token.AND_NOT_ASSIGN:
		return "andNotAssign()"
	case token.DEFINE:
		return "defineAssign()"
	case token.ASSIGN:
		return "assign()"
	default:
		return fmt.Sprintf("unknownAssign(\"%s\")", node.String())
	}
}

func branchTypeToRascal(node token.Token) string {
	switch node {
	case token.BREAK:
		return "breakBranch()"
	case token.CONTINUE:
		return "continueBranch()"
	case token.GOTO:
		return "gotoBranch()"
	case token.FALLTHROUGH:
		return "fallthroughBranch()"
	default:
		return fmt.Sprintf("unknownBranch(%s)", node.String())
	}
}

func declTypeToRascal(node token.Token) string {
	switch node {
	case token.IMPORT:
		return "importDecl()"
	case token.CONST:
		return "constDecl()"
	case token.TYPE:
		return "typeDecl()"
	case token.VAR:
		return "varDecl()"
	default:
		return fmt.Sprintf("unknownDecl(%s)", node.String())
	}
}

func labelToRascal(node *ast.Ident) string {
	if node != nil {
		return fmt.Sprintf("someLabel(\"%s\")", node.Name)
	} else {
		return "noLabel()"
	}
}

func rascalizeString(input string) string {
	//r := strings.NewReplacer("<", "\\<", ">", "\\>", "\n", "\\n", "\t", "\\t", "\r", "\\r", "\\", "\\\\", "\"", "\\\"", "'", "\\'")
	s1, _ := strings.CutPrefix(input, "\"")
	s2, _ := strings.CutSuffix(s1, "\"")
	return fmt.Sprintf("\"%s\"", rascalizer.Replace(s2))
}

func rascalizeChar(input string) string {
	//r := strings.NewReplacer("<", "\\<", ">", "\\>", "\n", "\\n", "\t", "\\t", "\r", "\\r", "\\", "\\\\", "\"", "\\\"", "'", "\\'")
	s1, _ := strings.CutPrefix(input, "'")
	s2, _ := strings.CutSuffix(s1, "'")
	return fmt.Sprintf("\"%s\"", rascalizer.Replace(s2))
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
