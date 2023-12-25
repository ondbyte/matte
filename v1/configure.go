package matte

import (
	"fmt"
	"go/ast"
	"go/format"
	"strings"
)

// configures your project
// this creates config.matte.go
func (m *Matte) Configure(wd string) error {
	_, err := GetPackagePathFromGoMod(wd)
	if err != nil {
		return err
	}
	return nil
}

func (m *Matte) StringifyTheAstNode(node ast.Node) (string, error) {
	strBuilder := strings.Builder{}
	err := format.Node(&strBuilder, m.fileSet, node)
	if err != nil {
		return "", err
	}
	return strBuilder.String(), nil
}

// gets the src string of the function used to configure the project
func (m *Matte) GetConfigFuncSrc() (string, error) {
	importStmt := &ast.ImportSpec{Path: &ast.BasicLit{Value: "github.com/ondbyte/matte/frameworks/config"}}
	corePkg := m.corePkg
	if corePkg == nil {
		return "", fmt.Errorf("corePkg should be non nil")
	}
	for _, file := range corePkg.Files {
		for _, i := range file.Imports {
			if i.Path.Value == importStmt.Path.Value {
				importStmt = i
			}
		}
		for _, decl := range file.Decls {
			decl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if decl.Name.Name == GeneralApiInfoFuncName {
				if importStmt == nil {
					return "", fmt.Errorf(
						`only a configuration function with name "%v" should exists, no other function with the name "%v" should exist`,
						GeneralApiInfoFuncName, GeneralApiInfoFuncName)
				}
				strAstNode, err := m.StringifyTheAstNode(decl)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("import %v\n\n%v", importStmt.Path.Value, strAstNode), nil
			}
		}
	}
	return "",
		fmt.Errorf(`a function named %v should be present in your project,
		which should return a slew of configurations specific to frameworks you use in your project.
		handlers will be parsed based on the configuration.
		for example,
		func %v() (frameworks.HttpConfig, frameworks.GinConfig, ...){
			...
		}
		`, GeneralApiInfoFuncName, GeneralApiInfoFuncName)
}

func (m *Matte) LoadConfig() error {
	/* funcSrc, err := m.GetConfigFuncSrc()
	if err != nil {
		return err
	}
	modCache := os.Getenv("GOMODCACHE")
	if modCache == "" {
		return fmt.Errorf("only go modules enabled projects are supported")
	}
	i := interp.New(interp.Options{GoPath: modCache})
	i.Use(std_lib.Symbol) */

	return nil
}
