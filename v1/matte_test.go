package matte_test

import (
	"fmt"
	"testing"

	"github.com/ondbyte/matte/v1"
)

func TestDemo(t *testing.T) {
	fmt.Println(matte.ParseDemoFunction())
}

func TestWithHttpFramework(t *testing.T) {
	var _ = `
package corehttp

import "net/http"

// @Router /hello [get]
func HandlePost(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ondbyte"))
}

// @Router /hello2 [get]
func HandleGet(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ondbyte"))
}
`
	/* assert := assert.New(t)
	framework, err := frameworks.NewHttp(":8888")
	assert.NoError(err)
	file, err := parser.ParseFile(token.NewFileSet(), "corehttp", errorSrc, parser.AllErrors|parser.ParseComments)
	assert.NoError(err)
	err = matte.ProcessFile(file, "corehttp", frameworks.Frameworks{framework})
	assert.NoError(err, "expected no error on ProcessFile but got : %v", err)
	data, err := framework.Finalize()
	assert.NoError(err, "finalize shouldnt have resulted in error")
	fmt.Println(data) */
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
