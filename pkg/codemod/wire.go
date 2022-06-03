package codemod

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func InsertWireModuleItems(filePath, items string) error {
	fileAST, err := parser.ParseFile(token.NewFileSet(), filePath, nil, 0)
	if err != nil {
		return fmt.Errorf("failed to parse file; %v", err)
	}

	for _, decl := range fileAST.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		if genDecl.Tok != token.VAR || len(genDecl.Specs) != 1 {
			continue
		}

		valSpec, ok := genDecl.Specs[0].(*ast.ValueSpec)
		if !ok {
			continue
		}

		if len(valSpec.Names) != 1 || valSpec.Names[0].String() != "WireModule" || len(valSpec.Values) != 1 {
			continue
		}

		offset := int(valSpec.End() - 2)

		return InsertTextToFile(filePath, items, offset)
	}

	return errors.New("failed to find WireModule declaration")
}
