package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

var rascalizer = strings.NewReplacer("<", "\\<", ">", "\\>", "\n", "\\n", "\t", "\\t", "\r", "\\r", "\\", "\\\\", "\"", "\\\"", "'", "\\'")
var filePath = ""
var projectPath = ""
var scratchDir = ""

type Config struct {
	addLocs  bool
	addTypes bool
	info     *types.Info
}

func addKeywordParameters(node ast.Node, fset *token.FileSet, config *Config) string {
	res := ""
	params := make(map[string]string)

	if config.addLocs {
		params["at"] = computeLocation(fset, node.Pos(), node.End())
	}

	if config.addTypes {
		if expr, ok := node.(ast.Expr); ok {
			if t, ok := config.info.Types[expr]; ok {
				params["exprType"] = rascalizeString(t.Type.String())
			}
		}
	}

	for key, value := range params {
		res = res + fmt.Sprintf(", %s=%s", key, value)
	}
	return res
}

// parses and type checks an entire project, returning a map of type-annotated AST nodes
func processProject(addLocs bool, addTypes bool, includeTests bool) string {
	// We want type-annotated syntax for the files in the current project
	cfg := &packages.Config{
		Mode:  packages.LoadSyntax,
		Dir:   projectPath,
		Tests: includeTests,
	}

	// Load the project, computing types and resolving dependencies
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Sprintf("errorSystem(\"Could not load project at path %s, errors: %s\")", projectPath, err.Error())
	}

	// Delete and recreate the scratch directory to make sure it's empty
	if err := os.RemoveAll(scratchDir); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(scratchDir, 0700); err != nil {
		panic(err)
	}

	// Generate Rascal terms for the files in the project package, saving
	// them each into a separate file in the scratch directory
	var files = make(map[string]string)
	counter := 0
	for _, pkg := range pkgs {
		config := &Config{
			addLocs:  addLocs,
			addTypes: addTypes,
			info:     pkg.TypesInfo,
		}
		for _, file := range pkg.Syntax {
			// Visit each file in the project/package
			fileAst := visitFile(file, pkg.Fset, config)
			filePath := filepath.Join(scratchDir, fmt.Sprintf("go-ast_%d.ast", counter))
			counter += 1
			if err := os.WriteFile(filePath, []byte(fileAst), 0644); err != nil {
				panic(err)
			}
			files[fmt.Sprintf("|file://%s|", pkg.Fset.Position(file.Pos()).Filename)] =
				fmt.Sprintf("|file://%s|", filePath)
		}
	}

	// Finally, generate the term representing the file map that we can use to load the
	// system in Rascal
	var filelist = make([]string, 0, len(files))
	for k, v := range files {
		filelist = append(filelist, fmt.Sprintf("%s : %s", k, v))
	}
	return fmt.Sprintf("(%s)", strings.Join(filelist, ", "))
}

// parses file and makes sure there is no error in file input.
func processFile(addLocs bool, addTypes bool) string {
	fset := token.NewFileSet()
	if file, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors); err != nil {
		return fmt.Sprintf("errorFile(Could not process file %s, %s)", filePath, err.Error())
	} else {
		conf := types.Config{
			Importer: importer.ForCompiler(fset, "source", nil),
			Error: func(err error) {
				fmt.Printf("Found a type error: %v\n", err)
			},
		}
		info := &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Defs:  make(map[*ast.Ident]types.Object),
			Uses:  make(map[*ast.Ident]types.Object),
		}
		if _, err := conf.Check(filePath, fset, []*ast.File{file}, info); err != nil {
			return fmt.Sprintf("errorFile(Could not process file %s, %s)", filePath, err.Error())
		}
		config := &Config{
			addLocs:  addLocs,
			addTypes: addTypes,
			info:     info,
		}
		return visitFile(file, fset, config)
	}
}

func processExpr(exprText string) string {
	fset := token.NewFileSet()
	if expr, err := parser.ParseExpr(exprText); err != nil {
		return fmt.Sprintf("unknownExpr(%s)", err.Error())
	} else {
		// TODO: See if we can create a checker just for this expression
		config := &Config{
			addLocs:  false,
			addTypes: false,
			info:     nil,
		}
		return visitExpr(&expr, fset, config)
	}
}

func visitFile(node *ast.File, fset *token.FileSet, config *Config) string {
	var decls []string
	for i := 0; i < len(node.Decls); i++ {
		decls = append(decls, visitDeclaration(&node.Decls[i], fset, config))
	}
	declString := strings.Join(decls, ",")

	// These are already part of the declarations...
	//var imports []string
	//for i := 0; i < len(node.Imports); i++ {
	//	imports = append(imports, visitImportSpec(node.Imports[i], fset, addLocs))
	//}
	//importString := strings.Join(imports, ",")

	packageName := node.Name.Name
	return fmt.Sprintf("file(\"%s\", [%s]%s)", packageName, declString, addKeywordParameters(node, fset, config))
}

func visitDeclaration(node *ast.Decl, fset *token.FileSet, config *Config) string {
	switch d := (*node).(type) {
	case *ast.GenDecl:
		return visitGeneralDeclaration(d, fset, config)
	case *ast.FuncDecl:
		return visitFunctionDeclaration(d, fset, config)
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

func visitSpec(node *ast.Spec, fset *token.FileSet, config *Config) string {
	switch d := (*node).(type) {
	case *ast.ImportSpec:
		return visitImportSpec(d, fset, config)
	case *ast.ValueSpec:
		return visitValueSpec(d, fset, config)
	case *ast.TypeSpec:
		return visitTypeSpec(d, fset, config)
	default:
		return "unknownSpec()" // This is an error, should panic here
	}
}

func visitImportSpec(node *ast.ImportSpec, fset *token.FileSet, config *Config) string {
	specName := optionalNameToRascal(node.Name)
	specPath := literalToRascal(node.Path, fset, config)
	return fmt.Sprintf("importSpec(%s,%s%s)", specName, specPath, addKeywordParameters(node, fset, config))
}

func visitValueSpec(node *ast.ValueSpec, fset *token.FileSet, config *Config) string {
	var names []string
	for i := 0; i < len(node.Names); i++ {
		names = append(names, fmt.Sprintf("\"%s\"", node.Names[i].Name))
	}
	namesStr := fmt.Sprintf("[%s]", strings.Join(names, ","))
	typeStr := visitOptionExpr(&node.Type, fset, config)
	values := visitExprList(node.Values, fset, config)

	return fmt.Sprintf("valueSpec(%s,%s,%s%s)", namesStr, typeStr, values, addKeywordParameters(node, fset, config))
}

func visitTypeSpec(node *ast.TypeSpec, fset *token.FileSet, config *Config) string {
	typeParams := visitFieldList(node.TypeParams, fset, config)
	typeStr := visitExpr(&node.Type, fset, config)

	return fmt.Sprintf("typeSpec(\"%s\",%s,%s%s)", node.Name.Name, typeParams, typeStr, addKeywordParameters(node, fset, config))
}

func visitSpecList(nodes []ast.Spec, fset *token.FileSet, config *Config) string {
	var specs []string
	for i := 0; nodes != nil && i < len(nodes); i++ {
		specs = append(specs, visitSpec(&nodes[i], fset, config))
	}
	return fmt.Sprintf("[%s]", strings.Join(specs, ","))
}

func visitGeneralDeclaration(node *ast.GenDecl, fset *token.FileSet, config *Config) string {
	declType := declTypeToRascal(node.Tok)
	specList := visitSpecList(node.Specs, fset, config)
	return fmt.Sprintf("genDecl(%s,%s%s)", declType, specList, addKeywordParameters(node, fset, config))
}

func visitFunctionDeclaration(node *ast.FuncDecl, fset *token.FileSet, config *Config) string {
	receivers := visitFieldList(node.Recv, fset, config)
	signature := visitFuncType(node.Type, fset, config)
	body := visitOptionalBlockStmt(node.Body, fset, config)
	return fmt.Sprintf("funDecl(\"%s\",%s,%s,%s%s)", node.Name, receivers, signature, body, addKeywordParameters(node, fset, config))
}

func visitStmt(node *ast.Stmt, fset *token.FileSet, config *Config) string {
	switch t := (*node).(type) {
	case *ast.DeclStmt:
		return visitDeclStmt(t, fset, config)
	case *ast.EmptyStmt:
		return visitEmptyStmt(t, fset, config)
	case *ast.LabeledStmt:
		return visitLabeledStmt(t, fset, config)
	case *ast.ExprStmt:
		return visitExprStmt(t, fset, config)
	case *ast.SendStmt:
		return visitSendStmt(t, fset, config)
	case *ast.IncDecStmt:
		return visitIncDecStmt(t, fset, config)
	case *ast.AssignStmt:
		return visitAssignStmt(t, fset, config)
	case *ast.GoStmt:
		return visitGoStmt(t, fset, config)
	case *ast.DeferStmt:
		return visitDeferStmt(t, fset, config)
	case *ast.ReturnStmt:
		return visitReturnStmt(t, fset, config)
	case *ast.BranchStmt:
		return visitBranchStmt(t, fset, config)
	case *ast.BlockStmt:
		return visitBlockStmt(t, fset, config)
	case *ast.IfStmt:
		return visitIfStmt(t, fset, config)
	case *ast.CaseClause:
		return visitCaseClause(t, fset, config)
	case *ast.SwitchStmt:
		return visitSwitchStmt(t, fset, config)
	case *ast.TypeSwitchStmt:
		return visitTypeSwitchStmt(t, fset, config)
	case *ast.CommClause:
		return visitCommClause(t, fset, config)
	case *ast.SelectStmt:
		return visitSelectStmt(t, fset, config)
	case *ast.ForStmt:
		return visitForStmt(t, fset, config)
	case *ast.RangeStmt:
		return visitRangeStmt(t, fset, config)
	default:
		return fmt.Sprintf("unknownStmt(\"%s\")", reflect.TypeOf(node).Name())
	}
}

func visitOptionStmt(node *ast.Stmt, fset *token.FileSet, config *Config) string {
	if *node == nil {
		return "noStmt()"
	} else {
		return fmt.Sprintf("someStmt(%s)", visitStmt(node, fset, config))
	}
}

func visitOptionExpr(node *ast.Expr, fset *token.FileSet, config *Config) string {
	if *node == nil {
		return "noExpr()"
	} else {
		return fmt.Sprintf("someExpr(%s)", visitExpr(node, fset, config))
	}
}

func visitDeclStmt(node *ast.DeclStmt, fset *token.FileSet, config *Config) string {
	declStr := visitDeclaration(&node.Decl, fset, config)
	return fmt.Sprintf("declStmt(%s)", declStr, addKeywordParameters(node, fset, config))
}

func visitEmptyStmt(node *ast.EmptyStmt, fset *token.FileSet, config *Config) string {
	return fmt.Sprintf("emptyStmt(%s)", addKeywordParameters(node, fset, config))
}

func visitLabeledStmt(node *ast.LabeledStmt, fset *token.FileSet, config *Config) string {
	labelStr := labelToRascal(node.Label)
	stmtStr := visitStmt(&node.Stmt, fset, config)
	return fmt.Sprintf("labeledStmt(%s,%s%s)", labelStr, stmtStr, addKeywordParameters(node, fset, config))
}

func visitExprStmt(node *ast.ExprStmt, fset *token.FileSet, config *Config) string {
	exprStr := visitExpr(&node.X, fset, config)
	return fmt.Sprintf("exprStmt(%s%s)", exprStr, addKeywordParameters(node, fset, config))
}

func visitSendStmt(node *ast.SendStmt, fset *token.FileSet, config *Config) string {
	chanStr := visitExpr(&node.Chan, fset, config)
	valStr := visitExpr(&node.Value, fset, config)
	return fmt.Sprintf("sendStmt(%s,%s%s)", chanStr, valStr, addKeywordParameters(node, fset, config))
}

func visitIncDecStmt(node *ast.IncDecStmt, fset *token.FileSet, config *Config) string {
	opStr := opToRascal(node.Tok)
	exprStr := visitExpr(&node.X, fset, config)

	return fmt.Sprintf("incDecStmt(%s,%s%s)", opStr, exprStr, addKeywordParameters(node, fset, config))
}

func visitAssignStmt(node *ast.AssignStmt, fset *token.FileSet, config *Config) string {
	assignOp := assignmentOpToRascal(node.Tok)
	left := visitExprList(node.Lhs, fset, config)
	right := visitExprList(node.Rhs, fset, config)
	return fmt.Sprintf("assignStmt(%s,%s,%s%s)", left, right, assignOp, addKeywordParameters(node, fset, config))
}

func visitGoStmt(node *ast.GoStmt, fset *token.FileSet, config *Config) string {
	exprStr := visitCallExpr(node.Call, fset, config)
	return fmt.Sprintf("goStmt(%s%s)", exprStr, addKeywordParameters(node, fset, config))
}

func visitDeferStmt(node *ast.DeferStmt, fset *token.FileSet, config *Config) string {
	exprStr := visitCallExpr(node.Call, fset, config)
	return fmt.Sprintf("deferStmt(%s%s)", exprStr, addKeywordParameters(node, fset, config))
}

func visitReturnStmt(node *ast.ReturnStmt, fset *token.FileSet, config *Config) string {
	results := visitExprList(node.Results, fset, config)
	return fmt.Sprintf("returnStmt(%s%s)", results, addKeywordParameters(node, fset, config))
}

func visitBranchStmt(node *ast.BranchStmt, fset *token.FileSet, config *Config) string {
	typeStr := branchTypeToRascal(node.Tok)
	labelStr := labelToRascal(node.Label)
	return fmt.Sprintf("branchStmt(%s,%s%s)", typeStr, labelStr, addKeywordParameters(node, fset, config))
}

func visitOptionalBlockStmt(node *ast.BlockStmt, fset *token.FileSet, config *Config) string {
	if node == nil {
		return "noStmt()"
	} else {
		return fmt.Sprintf("someStmt(%s)", visitBlockStmt(node, fset, config))
	}
}

func visitBlockStmt(node *ast.BlockStmt, fset *token.FileSet, config *Config) string {
	stmts := visitStmtList(node.List, fset, config)
	return fmt.Sprintf("blockStmt(%s%s)", stmts, addKeywordParameters(node, fset, config))
}

func visitIfStmt(node *ast.IfStmt, fset *token.FileSet, config *Config) string {
	initStmt := visitOptionStmt(&node.Init, fset, config)
	condExpr := visitExpr(&node.Cond, fset, config)
	body := visitBlockStmt(node.Body, fset, config)
	elseStmt := visitOptionStmt(&node.Else, fset, config)
	return fmt.Sprintf("ifStmt(%s,%s,%s,%s%s)", initStmt, condExpr, body, elseStmt, addKeywordParameters(node, fset, config))
}

func caseToRascal(nodes []ast.Expr, fset *token.FileSet, config *Config) string {
	if nodes != nil {
		return fmt.Sprintf("regularCase(%s)", visitExprList(nodes, fset, config))
	} else {
		return "defaultCase()"
	}
}

func visitCaseClause(node *ast.CaseClause, fset *token.FileSet, config *Config) string {
	caseStr := caseToRascal(node.List, fset, config)
	stmtsString := visitStmtList(node.Body, fset, config)

	return fmt.Sprintf("caseClause(%s,%s%s)", caseStr, stmtsString, addKeywordParameters(node, fset, config))
}

func visitCaseClauseList(node *ast.BlockStmt, fset *token.FileSet, config *Config) string {
	var cases []string
	for i := 0; node != nil && node.List != nil && i < len(node.List); i++ {
		nodeAsCase, ok := node.List[i].(*ast.CaseClause)
		if ok {
			cases = append(cases, visitCaseClause(nodeAsCase, fset, config))
		} else {
			cases = append(cases, "invalidCaseClause()")
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(cases, ","))
}

func visitSwitchStmt(node *ast.SwitchStmt, fset *token.FileSet, config *Config) string {
	init := visitOptionStmt(&node.Init, fset, config)
	tag := visitOptionExpr(&node.Tag, fset, config)
	block := visitCaseClauseList(node.Body, fset, config)
	return fmt.Sprintf("switchStmt(%s,%s,%s%s)", init, tag, block, addKeywordParameters(node, fset, config))
}

func visitTypeSwitchStmt(node *ast.TypeSwitchStmt, fset *token.FileSet, config *Config) string {
	init := visitOptionStmt(&node.Init, fset, config)
	assign := visitStmt(&node.Assign, fset, config)
	block := visitCaseClauseList(node.Body, fset, config)
	return fmt.Sprintf("typeSwitchStmt(%s,%s,%s%s)", init, assign, block, addKeywordParameters(node, fset, config))
}

func clauseToRascal(node *ast.Stmt, fset *token.FileSet, config *Config) string {
	if *node == nil {
		return "defaultComm()"
	} else {
		return fmt.Sprintf("regularComm(%s)", visitStmt(node, fset, config))
	}
}

func visitCommClause(node *ast.CommClause, fset *token.FileSet, config *Config) string {
	clauseStr := clauseToRascal(&node.Comm, fset, config)
	stmtsString := visitStmtList(node.Body, fset, config)

	return fmt.Sprintf("commClause(%s,%s%s)", clauseStr, stmtsString, addKeywordParameters(node, fset, config))
}

func visitCommClauseList(node *ast.BlockStmt, fset *token.FileSet, config *Config) string {
	var clauses []string
	for i := 0; node != nil && node.List != nil && i < len(node.List); i++ {
		nodeAsClause, ok := node.List[i].(*ast.CommClause)
		if ok {
			clauses = append(clauses, visitCommClause(nodeAsClause, fset, config))
		} else {
			clauses = append(clauses, "invalidCommClause()")
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(clauses, ","))
}

func visitSelectStmt(node *ast.SelectStmt, fset *token.FileSet, config *Config) string {
	clauses := visitCommClauseList(node.Body, fset, config)
	return fmt.Sprintf("selectStmt(%s%s)", clauses, addKeywordParameters(node, fset, config))
}

func visitForStmt(node *ast.ForStmt, fset *token.FileSet, config *Config) string {
	init := visitOptionStmt(&node.Init, fset, config)
	post := visitOptionStmt(&node.Post, fset, config)
	cond := visitOptionExpr(&node.Cond, fset, config)
	block := visitBlockStmt(node.Body, fset, config)
	return fmt.Sprintf("forStmt(%s,%s,%s,%s%s)", init, cond, post, block, addKeywordParameters(node, fset, config))
}

func visitRangeStmt(node *ast.RangeStmt, fset *token.FileSet, config *Config) string {
	key := visitOptionExpr(&node.Key, fset, config)
	value := visitOptionExpr(&node.Value, fset, config)
	assignOp := assignmentOpToRascal(node.Tok)
	rangeExpr := visitExpr(&node.X, fset, config)
	block := visitBlockStmt(node.Body, fset, config)
	return fmt.Sprintf("rangeStmt(%s,%s,%s,%s,%s%s)", key, value, assignOp, rangeExpr, block, addKeywordParameters(node, fset, config))
}

func visitExpr(node *ast.Expr, fset *token.FileSet, config *Config) string {
	switch t := (*node).(type) {

	case *ast.Ident:
		return visitIdent(t, fset, config)

	case *ast.Ellipsis:
		return visitEllipsis(t, fset, config)

	case *ast.BasicLit:
		return visitBasicLit(t, fset, config)

	case *ast.FuncLit:
		return visitFuncLit(t, fset, config)

	case *ast.CompositeLit:
		return visitCompositeLit(t, fset, config)

	case *ast.ParenExpr:
		return visitParenExpr(t, fset, config)

	case *ast.SelectorExpr:
		return visitSelectorExpr(t, fset, config)

	case *ast.IndexExpr:
		return visitIndexExpr(t, fset, config)

	case *ast.IndexListExpr:
		return visitIndexListExpr(t, fset, config)

	case *ast.SliceExpr:
		return visitSliceExpr(t, fset, config)

	case *ast.TypeAssertExpr:
		return visitTypeAssertExpr(t, fset, config)

	case *ast.CallExpr:
		return visitCallExpr(t, fset, config)

	case *ast.StarExpr:
		return visitStarExpr(t, fset, config)

	case *ast.UnaryExpr:
		return visitUnaryExpr(t, fset, config)

	case *ast.BinaryExpr:
		return visitBinaryExpr(t, fset, config)

	case *ast.KeyValueExpr:
		return visitKeyValueExpr(t, fset, config)

	case *ast.ArrayType:
		return visitArrayType(t, fset, config)

	case *ast.StructType:
		return visitStructType(t, fset, config)

	case *ast.FuncType:
		return visitFuncType(t, fset, config)

	case *ast.InterfaceType:
		return visitInterfaceType(t, fset, config)

	case *ast.MapType:
		return visitMapType(t, fset, config)

	case *ast.ChanType:
		return visitChanType(t, fset, config)

	default:
		return fmt.Sprintf("unknownExpr(\"%s\")", reflect.TypeOf(node).Name())
	}
}

func visitIdent(node *ast.Ident, fset *token.FileSet, config *Config) string {
	name := node.Name
	return fmt.Sprintf("ident(\"%s\"%s)", name, addKeywordParameters(node, fset, config))
}
func visitEllipsis(node *ast.Ellipsis, fset *token.FileSet, config *Config) string {
	elt := visitOptionExpr(&node.Elt, fset, config)
	return fmt.Sprintf("ellipsis(%s%s)", elt, addKeywordParameters(node, fset, config))
}

func literalToRascal(node *ast.BasicLit, fset *token.FileSet, config *Config) string {
	switch node.Kind {
	case token.INT:
		if parsed, err := strconv.ParseInt(node.Value, 0, 64); err == nil {
			return fmt.Sprintf("literalInt(%d%s)", parsed, addKeywordParameters(node, fset, config))
		} else if parsed, err := strconv.ParseUint(node.Value, 0, 64); err == nil {
			return fmt.Sprintf("literalInt(%d%s)", parsed, addKeywordParameters(node, fset, config))
		} else if parsed, err := strconv.ParseFloat(node.Value, 64); err == nil {
			// NOTE: This is here because we can include an INT literal as an argument
			// when setting up a floating-point number that is outside the possible
			// bounds of an int64, e.g.,340282346638528860000000000000000000000. This
			// is coerced into a float in that case, but cannot be represented as either
			// an int32 or an int64.
			return fmt.Sprintf("literalFloat(%f%s)", parsed, addKeywordParameters(node, fset, config))
		} else {
			return fmt.Sprintf("unknownLiteral(\"%s\"%s)", node.Value, addKeywordParameters(node, fset, config))
		}
	case token.FLOAT:
		if parsed, err := strconv.ParseFloat(node.Value, 64); err == nil {
			return fmt.Sprintf("literalFloat(%f)%s", parsed, addKeywordParameters(node, fset, config))
		} else {
			return fmt.Sprintf("unknownLiteral(\"%s\"%s)", node.Value, addKeywordParameters(node, fset, config))
		}
	case token.CHAR:
		return fmt.Sprintf("literalChar(%s%s)", rascalizeChar(node.Value), addKeywordParameters(node, fset, config))
	case token.STRING:
		return fmt.Sprintf("literalString(%s)%s", rascalizeString(node.Value), addKeywordParameters(node, fset, config))
	case token.IMAG:
		if ic, err := strconv.ParseComplex(node.Value, 128); err == nil {
			return fmt.Sprintf("literalImaginary(%f,%f%s)", real(ic), imag(ic), addKeywordParameters(node, fset, config))
		} else {
			return fmt.Sprintf("unknownLiteral(\"%s\"%s)", node.Value, addKeywordParameters(node, fset, config))
		}
	default:
		return fmt.Sprintf("unknownLiteral(\"%s\"%s)", node.Value, addKeywordParameters(node, fset, config))
	}
}

func visitBasicLit(node *ast.BasicLit, fset *token.FileSet, config *Config) string {
	value := literalToRascal(node, fset, config)
	return fmt.Sprintf("basicLit(%s%s)", value, addKeywordParameters(node, fset, config))
}

func visitFuncLit(node *ast.FuncLit, fset *token.FileSet, config *Config) string {
	typeStr := visitFuncType(node.Type, fset, config)
	body := visitBlockStmt(node.Body, fset, config)
	return fmt.Sprintf("funcLit(%s,%s%s)", typeStr, body, addKeywordParameters(node, fset, config))
}
func boolToRascal(val bool) string {
	if val {
		return "true"
	} else {
		return "false"
	}
}
func visitCompositeLit(node *ast.CompositeLit, fset *token.FileSet, config *Config) string {
	typeStr := visitOptionExpr(&node.Type, fset, config)
	elts := visitExprList(node.Elts, fset, config)

	return fmt.Sprintf("compositeLit(%s,%s,%s%s)", typeStr, elts, boolToRascal(node.Incomplete), addKeywordParameters(node, fset, config))
}

func visitParenExpr(node *ast.ParenExpr, fset *token.FileSet, config *Config) string {
	// We are building a tree, we do not need to keep explicit parens
	return visitExpr(&node.X, fset, config)
}

func visitSelectorExpr(node *ast.SelectorExpr, fset *token.FileSet, config *Config) string {
	exprStr := visitExpr(&node.X, fset, config)
	return fmt.Sprintf("selectorExpr(%s,\"%s\"%s)", exprStr, node.Sel.Name, addKeywordParameters(node, fset, config))
}
func visitIndexExpr(node *ast.IndexExpr, fset *token.FileSet, config *Config) string {
	x := visitExpr(&node.X, fset, config)
	index := visitExpr(&node.Index, fset, config)
	return fmt.Sprintf("indexExpr(%s,%s%s)", x, index, addKeywordParameters(node, fset, config))
}
func visitIndexListExpr(node *ast.IndexListExpr, fset *token.FileSet, config *Config) string {
	x := visitExpr(&node.X, fset, config)
	indices := visitExprList(node.Indices, fset, config)
	return fmt.Sprintf("indexListExpr(%s,%s%s)", x, indices, addKeywordParameters(node, fset, config))
}
func visitSliceExpr(node *ast.SliceExpr, fset *token.FileSet, config *Config) string {
	x := visitExpr(&node.X, fset, config)
	low := visitOptionExpr(&node.Low, fset, config)
	high := visitOptionExpr(&node.High, fset, config)
	maxnum := visitOptionExpr(&node.Max, fset, config)
	return fmt.Sprintf("sliceExpr(%s,%s,%s,%s,%s%s)", x, low, high, maxnum, boolToRascal(node.Slice3), addKeywordParameters(node, fset, config))
}
func visitTypeAssertExpr(node *ast.TypeAssertExpr, fset *token.FileSet, config *Config) string {
	x := visitExpr(&node.X, fset, config)
	exprTypes := visitOptionExpr(&node.Type, fset, config)
	return fmt.Sprintf("typeAssertExpr(%s,%s%s)", x, exprTypes, addKeywordParameters(node, fset, config))
}
func visitCallExpr(node *ast.CallExpr, fset *token.FileSet, config *Config) string {
	fun := visitExpr(&node.Fun, fset, config)
	args := visitExprList(node.Args, fset, config)
	hasEllipses := boolToRascal(node.Ellipsis != token.NoPos)
	return fmt.Sprintf("callExpr(%s,%s,%s%s)", fun, args, hasEllipses, addKeywordParameters(node, fset, config))
}
func visitStarExpr(node *ast.StarExpr, fset *token.FileSet, config *Config) string {
	x := visitExpr(&node.X, fset, config)
	return fmt.Sprintf("starExpr(%s%s)", x, addKeywordParameters(node, fset, config))
}
func visitUnaryExpr(node *ast.UnaryExpr, fset *token.FileSet, config *Config) string {
	x := visitExpr(&node.X, fset, config)
	tok := opToRascal(node.Op)
	return fmt.Sprintf("unaryExpr(%s,%s%s)", x, tok, addKeywordParameters(node, fset, config))
}

func visitBinaryExpr(node *ast.BinaryExpr, fset *token.FileSet, config *Config) string {
	x := visitExpr(&node.X, fset, config)
	tok := opToRascal(node.Op)
	y := visitExpr(&node.Y, fset, config)
	return fmt.Sprintf("binaryExpr(%s,%s,%s%s)", x, y, tok, addKeywordParameters(node, fset, config))
}
func visitKeyValueExpr(node *ast.KeyValueExpr, fset *token.FileSet, config *Config) string {
	key := visitExpr(&node.Key, fset, config)
	value := visitExpr(&node.Value, fset, config)
	return fmt.Sprintf("keyValueExpr(%s,%s%s)", key, value, addKeywordParameters(node, fset, config))
}
func visitArrayType(node *ast.ArrayType, fset *token.FileSet, config *Config) string {
	lens := visitOptionExpr(&node.Len, fset, config)
	elt := visitExpr(&node.Elt, fset, config)
	return fmt.Sprintf("arrayType(%s,%s%s)", lens, elt, addKeywordParameters(node, fset, config))
}

func visitOptionBasicLiteral(literal *ast.BasicLit, fset *token.FileSet, config *Config) string {
	if literal == nil {
		return "noLiteral()"
	} else {
		return fmt.Sprintf("someLiteral(%s)", literalToRascal(literal, fset, config))
	}
}

func fieldToRascal(field *ast.Field, fset *token.FileSet, config *Config) string {
	var names []string
	for i := 0; field.Names != nil && i < len(field.Names); i++ {
		names = append(names, fmt.Sprintf("\"%s\"", field.Names[i].Name))
	}
	namesStr := fmt.Sprintf("[%s]", strings.Join(names, ","))
	fieldType := visitOptionExpr(&field.Type, fset, config)
	fieldTag := visitOptionBasicLiteral(field.Tag, fset, config)

	return fmt.Sprintf("field(%s,%s,%s%s)", namesStr, fieldType, fieldTag, addKeywordParameters(field, fset, config))
}

func visitFieldList(fieldList *ast.FieldList, fset *token.FileSet, config *Config) string {
	var fields []string
	for i := 0; fieldList != nil && i < len(fieldList.List); i++ {
		fields = append(fields, fieldToRascal(fieldList.List[i], fset, config))
	}
	return fmt.Sprintf("[%s]", strings.Join(fields, ","))
}

func visitStructType(node *ast.StructType, fset *token.FileSet, config *Config) string {
	fieldStr := visitFieldList(node.Fields, fset, config)
	return fmt.Sprintf("structType(%s%s)", fieldStr, addKeywordParameters(node, fset, config))
}
func visitFuncType(node *ast.FuncType, fset *token.FileSet, config *Config) string {
	typeParams := visitFieldList(node.TypeParams, fset, config)
	params := visitFieldList(node.Params, fset, config)
	returns := visitFieldList(node.Results, fset, config)

	return fmt.Sprintf("funcType(%s,%s,%s%s)", typeParams, params, returns, addKeywordParameters(node, fset, config))
}
func visitInterfaceType(node *ast.InterfaceType, fset *token.FileSet, config *Config) string {
	methods := visitFieldList(node.Methods, fset, config)
	return fmt.Sprintf("interfaceType(%s%s)", methods, addKeywordParameters(node, fset, config))
}
func visitMapType(node *ast.MapType, fset *token.FileSet, config *Config) string {
	key := visitExpr(&node.Key, fset, config)
	value := visitExpr(&node.Value, fset, config)
	return fmt.Sprintf("mapType(%s,%s%s)", key, value, addKeywordParameters(node, fset, config))
}

func channelDirToRascal(dir ast.ChanDir) string {
	if (dir&ast.SEND == ast.SEND) && (dir&ast.RECV == ast.RECV) {
		return "bidirectional()"
	} else if dir&ast.SEND == ast.SEND {
		return "send()"
	} else if dir&ast.RECV == ast.RECV {
		return "receive()"
	} else {
		return "unknown()"
	}
}

func visitChanType(node *ast.ChanType, fset *token.FileSet, config *Config) string {
	value := visitExpr(&node.Value, fset, config)
	chanSend := channelDirToRascal(node.Dir)
	return fmt.Sprintf("chanType(%s,%s%s)", value, chanSend, addKeywordParameters(node, fset, config))
}

func visitExprList(nodes []ast.Expr, fset *token.FileSet, config *Config) string {
	var exprs []string
	for i := 0; nodes != nil && i < len(nodes); i++ {
		exprs = append(exprs, visitExpr(&nodes[i], fset, config))
	}
	return fmt.Sprintf("[%s]", strings.Join(exprs, ","))
}

func visitStmtList(nodes []ast.Stmt, fset *token.FileSet, config *Config) string {
	var stmts []string
	for i := 0; nodes != nil && i < len(nodes); i++ {
		stmts = append(stmts, visitStmt(&nodes[i], fset, config))
	}

	return fmt.Sprintf("[%s]", strings.Join(stmts, ","))
}

// computes location of the ast nodes
func computeLocation(fset *token.FileSet, start token.Pos, end token.Pos) string {
	if start.IsValid() && end.IsValid() {
		startPos := fset.Position(start)
		//endPos := fset.Position(end)
		return fmt.Sprintf("|file://%s|(%d,%d)",
			fset.Position(start).Filename, startPos.Offset, end-start)
		//startPos.Line, startPos.Column-1, endPos.Line, endPos.Column-1)
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
	case token.TILDE:
		return "tilde()"
	default:
		return fmt.Sprintf("unknownOp(\"%s\")", node.String())
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
	case token.ILLEGAL:
		return "noKey()"
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
	flag.StringVar(&filePath, "filePath", "", "The file to be processed")
	flag.StringVar(&projectPath, "projectPath", "", "The root of the project")
	flag.StringVar(&scratchDir, "scratchDir", "", "The directory for temporary AST files")

	var addLocations bool
	flag.BoolVar(&addLocations, "addLocs", true, "Include location annotations")

	var addTypes bool
	flag.BoolVar(&addTypes, "addTypes", true, "Include type information")

	var includeTests bool
	flag.BoolVar(&includeTests, "includeTests", false, "Include test files")

	var exprText string
	flag.StringVar(&exprText, "expr", "", "The expression to parse")

	flag.Parse()

	if projectPath != "" {
		if scratchDir != "" {
			fmt.Println(processProject(addLocations, addTypes, includeTests))
		} else {
			fmt.Println("No scratch directory given for generated ASTs")
		}
	} else if filePath != "" {
		//fmt.Printf("Processing file %s\n", filePath)
		fmt.Println(processFile(addLocations, addTypes))
	} else if exprText != "" {
		fmt.Println(processExpr(exprText))
	} else {
		fmt.Println("No file or expression given")
	}
}
