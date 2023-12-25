package frameworks

import (
	"errors"
	"go/ast"
)

var errIsNotHandler = errors.New("is not a handler")

type Config struct {
	HttpConfig HttpConfig
}

// represents a handler source we need to generate
type Handler struct {
	HttpMethod, Path, Handler string
}

type Framework interface {
	// check whether passed nfunction is a handler which is specific to a framework like gin
	IsHandler(fn *ast.FuncType) error
	//mounts the handler
	AddHandler(httpMethod, path, pkgImportPath, pkgName, handlerName string) error

	// returns the final src to write to file
	Finalize() (importStmt, funcName, funcSrc string, err error)

	// configure this framework
	Configure(configStruct interface{}) error
}

type Frameworks []Framework

func (f Frameworks) GetFrameworkForHandler(funcSignature *ast.FuncType) Framework {
	for _, f := range f {
		if err := f.IsHandler(funcSignature); err == nil {
			return f
		}
	}
	return nil
}

func contains(slice []interface{}, value interface{}) bool {
	for _, element := range slice {
		if element == value {
			return true
		}
	}
	return false
}
