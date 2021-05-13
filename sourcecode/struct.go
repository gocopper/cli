package sourcecode

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"

	"github.com/gocopper/copper/cerrors"
)

type Struct struct {
	Name   string
	Fields []StructField
}

type StructField struct {
	Name string
	Type string
}

func FindStructs(filePath string) ([]Struct, error) {
	fs := token.NewFileSet()

	fileAST, err := parser.ParseFile(fs, filePath, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file; %v", err)
	}

	structs := make([]Struct, 0)

	for _, decl := range fileAST.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		if genDecl.Tok != token.TYPE {
			continue
		}

		if len(genDecl.Specs) != 1 {
			continue
		}

		typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)
		if !ok {
			continue
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		newStruct := Struct{
			Name:   typeSpec.Name.String(),
			Fields: make([]StructField, len(structType.Fields.List)),
		}

		for i, f := range structType.Fields.List {
			var typeSb strings.Builder

			if len(f.Names) != 1 {
				continue
			}

			err = printer.Fprint(&typeSb, fs, f.Type)
			if err != nil {
				return nil, cerrors.New(err, "failed to get struct's field type string representation", map[string]interface{}{
					"structName": newStruct.Name,
					"field":      f.Names[0].String(),
				})
			}

			newStruct.Fields[i] = StructField{
				Name: f.Names[0].String(),
				Type: typeSb.String(),
			}
		}

		structs = append(structs, newStruct)
	}

	return structs, nil
}

func FindStructMethodNames(filePath, structName string) ([]string, error) {
	methods := make([]string, 0)

	fileAST, err := parser.ParseFile(token.NewFileSet(), filePath, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file; %v", err)
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

		if ident.Name != structName {
			continue
		}

		methods = append(methods, funcDecl.Name.String())
	}

	return methods, nil
}
