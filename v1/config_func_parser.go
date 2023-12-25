package matte

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/ondbyte/matte/frameworks"
)

func demo() (frameworks.HttpConfig, frameworks.HttpConfig) {
	return frameworks.HttpConfig{
			Addr: "01",
		},
		frameworks.HttpConfig{
			Addr: "02",
		}
}

func ParseDemoFunction() (map[string]map[string]string, error) {
	src := `
package main

import  "github.com/ondbyte/matte/frameworks"

func demo() (frameworks.HttpConfig, frameworks.HttpConfig) {
	return frameworks.HttpConfig{
			Addr: "01",
		},
		frameworks.HttpConfig{
			Addr: "02",
		}
}
`

	fset := token.NewFileSet()

	// Parse the source code
	f, err := parser.ParseFile(fset, "demo.go", src, 0)
	if err != nil {
		return nil, fmt.Errorf("Error parsing source code:%v", err)
	}
	imprtName := ""
	for _, i := range f.Imports {
		if i.Path.Value == "\"github.com/ondbyte/matte/frameworks\"" {
			if i.Name == nil {
				imprtName = "frameworks"
				break
			}
			imprtName = i.Name.Name
		}
	}
	if imprtName == "" {
		return nil, fmt.Errorf("invalid import")
	}

	// Find the demo function
	var configFunction *ast.FuncDecl
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "demo" {
			configFunction = fn
			break
		}
	}

	configs := map[string]map[string]string{}
	// Extract struct values
	if configFunction != nil {
		for _, stmt := range configFunction.Body.List {
			retStmt, ok := stmt.(*ast.ReturnStmt)
			if !ok {
				return nil, fmt.Errorf("expected a return statement")
			}
			for _, result := range retStmt.Results {
				compLit, ok := result.(*ast.CompositeLit)
				if !ok {
					return nil, fmt.Errorf("expected a composite literal")
				}
				se, ok := compLit.Type.(*ast.SelectorExpr)
				configName := ""
				if !ok {
					justName, ok := compLit.Type.(*ast.Ident)
					if !ok {
						return nil, fmt.Errorf("expected a selector expression/an ident")
					}
					configName = justName.Name
				} else if se.Sel != nil {
					configName = se.Sel.Name
					ide, ok := se.X.(*ast.Ident)
					if !ok || ide.Name != imprtName {
						return nil, fmt.Errorf("expected identifier with name %v", imprtName)
					}
				} else {
					return nil, fmt.Errorf("selector cannot be nil")
				}
				configs[configName] = map[string]string{}
				for _, elt := range compLit.Elts {
					keyVal, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						return nil, fmt.Errorf("expected key value expected")
					}
					k, ok := keyVal.Key.(*ast.Ident)
					if !ok {
						return nil, fmt.Errorf("expected a ident on key side")
					}
					key := k.Name
					value, ok := keyVal.Value.(*ast.BasicLit)
					if !ok {
						return nil, fmt.Errorf("expected a string value on the value side of struct field %v", key)
					}
					configs[configName][key] = value.Value
				}
			}
		}
	}
	return configs, nil
}
