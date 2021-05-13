package sourcecode

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func PosToInsertNewRouteDecl(path string) (int, error) {
	const (
		routerStructName = "Router"
		routesMethodName = "Routes"
	)

	fileAST, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to parse file; %v", err)
	}

	for _, decl := range fileAST.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if funcDecl.Recv == nil || funcDecl.Recv.NumFields() != 1 {
			continue
		}

		startExpr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			continue
		}

		ident, ok := startExpr.X.(*ast.Ident)
		if !ok {
			continue
		}

		if ident.Name != routerStructName {
			continue
		}

		if funcDecl.Name.String() != routesMethodName {
			continue
		}

		if funcDecl.Body == nil {
			continue
		}

		for _, stmt := range funcDecl.Body.List {
			returnStmt, ok := stmt.(*ast.ReturnStmt)
			if !ok {
				continue
			}

			if len(returnStmt.Results) != 1 {
				continue
			}

			// check if routes are being returned as an array literal i.e. declared inline
			compositeLit, ok := returnStmt.Results[0].(*ast.CompositeLit)
			if !ok {
				continue
			}

			return int(compositeLit.Rbrace - 1), nil
		}
	}

	return 0, errors.New("failed to find router decl")
}
