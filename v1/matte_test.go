package matte_test

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/ondbyte/matte/v1"
	"github.com/stretchr/testify/assert"
	a "github.com/stretchr/testify/assert"
)

func TestParseParam(t *testing.T) {
	expr, _ := parser.ParseExpr(`func(a string,x byte,b uint,c *int){}`)
	flit, ok := expr.(*ast.FuncLit)
	assert.True(t, ok)
	for _, f := range flit.Type.Params.List {
		params, err := matte.ParseParam(f)
		if assert.NoError(t, err) {
			for _, p := range params {
				fmt.Println(p)
			}
		}
	}
}

func TestGetParamSVerifier(t *testing.T) {
	s := matte.GetParamsVerifierSrc([]*matte.Param{
		{
			Name:     "yadu",
			Type:     "*int",
			Required: false,
		},
		{
			Name:     "chinmaya",
			Type:     "uint",
			Required: true,
		},
		{
			Name:     "yadu2",
			Type:     "*int",
			Required: false,
		},
	}, "yadu.HandleHello")
	sb, err := format.Source([]byte(s))
	if assert.NoError(t, err) {
		fmt.Println(string(sb))
	}
}

func TestBuild(t *testing.T) {
	assert := a.New(t)
	dir, err := filepath.Abs("../test_project")
	assert.NoError(err)
	err = matte.Build(token.NewFileSet(), dir)
	if !assert.NoError(err) {
		return
	}
}

func TestWithHttpFrameworkHavingDuplicatePath(t *testing.T) {
	var _ = `
package corehttp

import "net/http"

// @Router /hello [get]
func HandlePost(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ondbyte"))
}

// @Router /hello [get]
func HandleGet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ondbyte"))
}
`
	/* assert := assert.New(t)
	_, httpFramework, err := frameworks.NewHttp(":8888")
	assert.NoError(err)
	file, err := parser.ParseFile(token.NewFileSet(), "corehttp", errorSrc, parser.AllErrors|parser.ParseComments)
	assert.NoError(err)
	err = matte.ProcessFile(file, "yadu", frameworks.Frameworks{httpFramework})
	expectedErr := `path /hello is already registered for handler yadu.HandlePost so cannot register it with handler yadu.HandleGet again`
	assert.Error(err, "expected a error")
	assert.Equal(errors.New(expectedErr), err, "expected error :%v", expectedErr) */
}
