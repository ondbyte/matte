package frameworks

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type HttpConfig struct {
	Addr string
}
type HttpFramework struct {
	HttpConfig
	funcBody      *strings.Builder
	imports       map[string]bool
	handlerSign   *ast.FuncType
	consumedPaths map[string]string
}

func (f *HttpFramework) LoadScaffoldCode(dir string) (string, error) {
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("unable read dir %v because err: %v", dir, err)
	}
	fileSet := token.NewFileSet()
	var files []*ast.File
	for _, filePath := range dirs {
		fileAST, err := parser.ParseFile(fileSet, filepath.Join(dir, filePath.Name()))
		if err != nil {
			return "", fmt.Errorf("error parsing file %s: %w", filePath, err)
		}
		files = append(files, fileAST)
	}

	// Merge the ASTs
	mergedAST := mergeASTs(files)

	// Format the merged AST to source code
	src, err := format.Source(mergedAST)
	if err != nil {
		return "", fmt.Errorf("error formatting merged AST: %w", err)
	}

	return string(src), nil
}

// Finalize implements Framework.
func (f *HttpFramework) Finalize() (importStmt, funcName, funcSrc string, err error) {
	_, p, _, _ := runtime.Caller(0)
	fmt.Println(p)
	importStmt = `"net/http"`
	for k, _ := range f.imports {
		importStmt += fmt.Sprintf("\n\"%v\"", k)
	}
	importStmt = fmt.Sprintf(`
	import (
		%v
	)
	`, importStmt)
	funcName = `RunHttpServer`
	body := fmt.Sprintf("func %v(){", funcName)
	body += f.funcBody.String()
	body += `err:=http.ListenAndServe("%v",http.DefaultServeMux)
	if err!=nil{
		panic(err)
	}
}`
	return importStmt, funcName, body, nil
}

// Configure implements Framework.
func (f *HttpFramework) Configure(configuration interface{}) error {
	cfg, ok := configuration.(HttpConfig)
	if !ok {
		return fmt.Errorf("configuration should be of type %+v", &HttpConfig{})
	}
	f.HttpConfig = cfg
	return nil
}

func NewHttp() Framework {
	handlerSignature := `func (w http.ResponseWriter,r *http.Request){}`
	x, _ := parser.ParseExpr(handlerSignature)
	return &HttpFramework{
		handlerSign:   x.(*ast.FuncLit).Type,
		funcBody:      &strings.Builder{},
		consumedPaths: map[string]string{},
		imports:       map[string]bool{},
	}
}

// AddHandler implements Framework.
func (f *HttpFramework) AddHandler(httpMethod string, path, importPath string, pkgName string, handlerName string) error {
	pkgHandler := fmt.Sprintf("%v.%v", pkgName, handlerName)
	prevHandler, ok := f.consumedPaths[path]
	if ok {
		return fmt.Errorf("path %v is already registered for handler %v so cannot register it with handler %v again", path, prevHandler, pkgHandler)
	}
	f.imports[importPath] = true
	_, err := f.funcBody.WriteString(fmt.Sprintf(`
	http.HandleFunc("%v",%v)%v`, path, pkgHandler, "\n"))
	if err != nil {
		return fmt.Errorf("cannot write to main string builder: %v", err)
	}
	f.consumedPaths[path] = pkgHandler
	return nil
}

// IsHandler implements Framework.
func (f *HttpFramework) IsHandler(fnType *ast.FuncType) error {
	if fnType == nil {
		return fmt.Errorf("fn cannot be nil")
	}
	handler := f.handlerSign
	a := handler.Params.List[0].Type.(*ast.SelectorExpr)
	b := handler.Params.List[1].Type.(*ast.StarExpr).X.(*ast.SelectorExpr)

	if len(fnType.Params.List) != 2 {
		return fmt.Errorf("handler should take two params")
	}
	x, ok := fnType.Params.List[0].Type.(*ast.SelectorExpr)
	if !ok || x.Sel == nil {
		return fmt.Errorf("first param of the handler should be a %v", a.Sel.Name)
	}
	if x.Sel.Name != a.Sel.Name {
		return fmt.Errorf("first param of the handler should be a %v but its %v", a.Sel.Name, x.Sel.Name)
	}
	starExp, ok := fnType.Params.List[1].Type.(*ast.StarExpr)
	y, ok2 := starExp.X.(*ast.SelectorExpr)
	if !ok || !ok2 {
		return fmt.Errorf("second param of the handler should be a %v", b.Sel.Name)
	}
	if y.Sel.Name != b.Sel.Name {
		return fmt.Errorf("second param of the handler should be a %v but its %v", b.Sel.Name, y.Sel.Name)
	}
	return nil
}
