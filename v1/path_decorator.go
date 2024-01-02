package matte

import (
	"fmt"
	"go/ast"
	"net/http"
	"strings"
)

type HttpMethod string

const (
	GET     HttpMethod = "GET"
	POST               = "POST"
	PUT                = "PUT"
	DELETE             = "DELETE"
	PATCH              = "PATCH"
	OPTIONS            = "OPTIONS"
)

type Path struct {
	path   string
	method HttpMethod
}

// isValidHTTPMethod checks if a given string is a valid HTTP method
func isValidHTTPMethod(method string) bool {
	// Convert the method to uppercase for case-insensitive comparison
	method = strings.ToUpper(method)

	// Check if the method is one of the allowed HTTP methods
	switch method {
	case "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS":
		return true
	default:
		return false
	}
}

func GetParamsVerifierSrc(params []*Param, caller string) string {
	s := "var err error"
	newLine := func() {
		s += "\n"
	}
	newLine()
	s += fmt.Sprintf(`errS:=""`)
	newLine()
	args := ""
	for _, param := range params {
		paramName := param.Name
		paramType := param.Type
		paramTypeWithoutStar := strings.Trim(paramType, "*")
		s += fmt.Sprintf(`%vS:=p.ByName("%v")`, paramName, paramName)
		newLine()
		if param.Required {
			s += fmt.Sprintf(`if %vS == "" {`, paramName)
			newLine()
			s += fmt.Sprintf(`errS += fmt.Sprintf("param '%v' is required\n")
			}`, paramName)
			newLine()
		}
		s += fmt.Sprintf(`%v:=new(%v)`, paramName, paramTypeWithoutStar)
		newLine()
		s += fmt.Sprintf(`if %vS != "" {`, paramName)
		newLine()
		s += fmt.Sprintf(`err=json.Unmarshal([]byte(%vS), %v)`, paramName, paramName)
		newLine()
		s += fmt.Sprintf(`if err!=nil{
			errS+="param value " + %vS +" cannot be Unmarshalled into type %v\n"
		}
		}`, paramName, paramTypeWithoutStar)
		newLine()
		s += fmt.Sprintf(`if errS!=""{
			http.Error(w,errS,http.StatusTeapot)
			return
		}`)
		newLine()
		if param.Required {
			args += "*"
		}
		args += paramName + ","
	}
	s += fmt.Sprintf("%v(%v)", caller, args)
	newLine()
	return s
}

func a(w http.ResponseWriter) {
	http.Error(w, "", http.StatusTeapot)
}

func (m *Matte) ProcessPath(
	pathDecorator *Decorator,
	handler *ast.FuncDecl,
) (err error) {
	caller := m.currentPkg.Name + "." + handler.Name.Name
	if len(pathDecorator.args) != 2 {
		err = fmt.Errorf("path decorator must have two args")
		return
	}
	httpMethod, path := strings.Trim(pathDecorator.args[0], `"`), pathDecorator.args[1]
	if !isValidHTTPMethod(httpMethod) {
		err = fmt.Errorf("invalid httpMethod")
		return
	}
	params := []*Param{}
	for _, param := range handler.Type.Params.List {
		_params, err := ParseParam(param)
		if err != nil {
			return fmt.Errorf("err while ParseForm : \n%v", err)
		}
		params = append(params, _params...)
	}
	verifierSrc := GetParamsVerifierSrc(params, caller)

	src := fmt.Sprintf(`
	router.Handle("%v",%v,func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		%v
	})
	`, httpMethod, path, verifierSrc)
	m.src += src
	return nil
}

type Param struct {
	Name     string
	Type     string
	Required bool
}

func ParseParam(field *ast.Field) (params []*Param, err error) {
	params = []*Param{}
	var id *ast.Ident
	var paramType string
	starExpr, ok := field.Type.(*ast.StarExpr)
	required := true
	if ok {
		paramType += "*"
		required = false
		id, _ = starExpr.X.(*ast.Ident)
	} else {
		id, _ = field.Type.(*ast.Ident)
	}
	if id == nil {
		return nil, fmt.Errorf("invalid type of param %v", field)
	}
	paramType += id.Name
	for _, name := range field.Names {
		params = append(params, &Param{Name: name.Name, Type: paramType, Required: required})
	}
	return params, nil
}

/*
func (o *OA) ParsePathDecorator(name string, args []string) error {
	decorator, _ := ParseDecorator(`path("GET","yadhu")`)
	if len(nil) != len(args) {
		return fmt.Errorf("needs two args")
	}
	if len(args) != 2 {
		return fmt.Errorf("path decorator needs two arguments")
	}
	if !isValidHTTPMethod(args[0]) {
		return fmt.Errorf("first argument must be a valid http method")
	}
	i := o.Paths.Find(args[1])
	if i == nil {
		i = &openapi3.PathItem{}
	}
	switch args[0] {
	case "GET":
		{
			if i.Get != nil {
				return fmt.Errorf("path '%v' already has a handler for method type '%v'", args[1], args[0])
			}
			op := openapi3.NewOperation()
			i.Get = op
		}

	}
	return nil
}
*/
