package webgo

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/gocopper/cli/v3/pkg/codemod"
	"github.com/gocopper/copper/cerrors"
)

func insertHTMLRouterToAppHandler(path string) error {
	fileAST, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return fmt.Errorf("failed to parse file; %v", err)
	}

	didInsertRouterParam := false

	for _, decl := range fileAST.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		if genDecl.Tok != token.TYPE || len(genDecl.Specs) != 1 {
			continue
		}

		typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)
		if !ok || typeSpec.Name.String() != "NewHTTPHandlerParams" {
			continue
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		offset := int(structType.Fields.Opening)

		paramText := fmt.Sprintf("\nHTML *chttp.HTMLRouter")

		err = codemod.InsertTextToFile(path, paramText, offset)
		if err != nil {
			return cerrors.New(err, "failed to add router param to NewHandlerParams", nil)
		}

		didInsertRouterParam = true
	}

	if !didInsertRouterParam {
		return errors.New("failed to find NewHandlerParams decl")
	}

	fileAST, err = parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return fmt.Errorf("failed to parse file; %v", err)
	}

	didInsertRouter := false

	for _, decl := range fileAST.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name.String() != "NewHTTPHandler" {
			continue
		}

		for _, stmt := range funcDecl.Body.List {
			retStmt, ok := stmt.(*ast.ReturnStmt)
			if !ok || len(retStmt.Results) != 1 {
				continue
			}

			callExpr, ok := retStmt.Results[0].(*ast.CallExpr)
			if !ok || len(callExpr.Args) != 1 {
				continue
			}

			compositeLit, ok := callExpr.Args[0].(*ast.CompositeLit)
			if !ok {
				continue
			}

			for _, elt := range compositeLit.Elts {
				kvExpr, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				keyIdent, ok := kvExpr.Key.(*ast.Ident)
				if !ok || keyIdent.String() != "Routers" {
					continue
				}

				valCompositeLit, ok := kvExpr.Value.(*ast.CompositeLit)
				if !ok {
					continue
				}

				offset := int(valCompositeLit.Lbrace)

				routerText := fmt.Sprintf("\np.HTML,\n")

				err = codemod.InsertTextToFile(path, routerText, offset)
				if err != nil {
					return cerrors.New(err, "failed to add router to []chttp.Router", nil)
				}

				didInsertRouter = true
			}
		}
	}

	if !didInsertRouter {
		return errors.New("failed to find []chttp.Router decl")
	}

	return nil
}
