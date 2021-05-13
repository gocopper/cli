package sourcecode

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"

	"github.com/gocopper/copper/cerrors"
	"github.com/iancoleman/strcase"
)

func InsertRouterToAppHandler(pkg string) error {
	const handlerFilePath = "pkg/app/handler.go"

	fileAST, err := parser.ParseFile(token.NewFileSet(), handlerFilePath, nil, 0)
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
		if !ok || typeSpec.Name.String() != "NewHandlerParams" {
			continue
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		offset := int(structType.Fields.Opening)

		paramText := fmt.Sprintf("\n%s *%s.Router", strcase.ToCamel(path.Base(pkg)), path.Base(pkg))

		err = insertText(handlerFilePath, paramText, offset)
		if err != nil {
			return cerrors.New(err, "failed to add router param to NewHandlerParams", nil)
		}

		didInsertRouterParam = true
	}

	if !didInsertRouterParam {
		return errors.New("failed to find NewHandlerParams decl")
	}

	fileAST, err = parser.ParseFile(token.NewFileSet(), handlerFilePath, nil, 0)
	if err != nil {
		return fmt.Errorf("failed to parse file; %v", err)
	}

	didInsertRouter := false

	for _, decl := range fileAST.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name.String() != "NewHandler" {
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

				routerText := fmt.Sprintf("\np.%s,", strcase.ToCamel(path.Base(pkg)))

				err = insertText(handlerFilePath, routerText, offset)
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
