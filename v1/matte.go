package matte

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/rogpeppe/go-internal/modfile"
)

type Pkg struct {
	*ast.Package
	ImportPath string
	Name       string
}

type Matte struct {
	matteDir string
	fileSet  *token.FileSet
	wd       string
	corePkg  *Pkg
	// refers the pkg which is being processed
	currentPkg *Pkg
	Pkgs       []*Pkg
	modFile    *modfile.File
	src        string
	imports    string
}

const MatteDir = "matte"

func createMatteDir(projectDir string) (string, error) {
	projectDir = filepath.Clean(projectDir)
	dir := filepath.Join(projectDir, MatteDir)
	stat, err := os.Stat(dir)
	if stat != nil && err == nil {
		err := os.RemoveAll(dir)
		if err != nil {
			return "", fmt.Errorf("unable to delete matte dir due to err: %v", err)
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("error while checking if the matteDir already exists due to err: %v", err)
	}
	err = os.Mkdir(dir, 0777)
	if err != nil {
		return "", fmt.Errorf("unable to mkdir %v due to err: %v", dir, err)
	}
	return dir, nil
}

// builds the project at path 'project'
func Build(fileSet *token.FileSet, project string) error {
	matteDir, err := createMatteDir(project)
	if err != nil {
		return fmt.Errorf("unable to make matteDir due to err: %v", err)
	}
	m := &Matte{
		fileSet:  fileSet,
		wd:       project,
		matteDir: matteDir,
	}
	// defer clean up
	//defer m.DeferCleanUp()
	err = m.parseModFile()
	if err != nil {
		return err
	}
	err = m.loadProject()
	if err != nil {
		return err
	}
	err = m.processProject()
	if err != nil {
		return err
	}
	err = m.build()
	return err
}

func (m *Matte) build() error {
	srcS := fmt.Sprintf(`
	package main

	import (
		"encoding/json"
		"fmt"
		"net/http"
	
		"github.com/julienschmidt/httprouter"
		%v
	)
	func main(){
		router := httprouter.New()
		%v
	}
	`, m.imports, m.src)
	_, err := format.Source([]byte(srcS))
	if err != nil {
		//return fmt.Errorf("failed to format go src due to err: %v", err)
	}
	err = os.WriteFile(filepath.Join(m.matteDir, "app.go"), []byte(srcS), 0777)
	if err != nil {
		return fmt.Errorf("failed to write file app.go due to err: %v", err)
	}
	return nil
}

func (m *Matte) parseModFile() error {
	if m.modFile != nil {
		return fmt.Errorf("go.mod file has been parsed already")
	}
	goModFilePath := filepath.Join(m.wd, "go.mod")
	goModData, err := os.ReadFile(goModFilePath)
	if err != nil {
		if os.IsNotExist(err) && m.modFile == nil {
			return fmt.Errorf("your project dir '%v' must contain a go.mod file, but it does not", m.wd)
		}
		return fmt.Errorf("failed to read the go mod file at %v because err: %v", goModFilePath, err)
	}
	modFile, err := modfile.Parse(goModFilePath, goModData, nil)
	if err != nil {
		return fmt.Errorf("unable to parse mod file data, file is at %v", goModFilePath)
	}
	m.modFile = modFile
	return nil
}

func (m *Matte) loadProject() (first error) {
	flags := parser.AllErrors | parser.ParseComments
	m.Pkgs = []*Pkg{}
	pkg, dirs, err := m.parseDir(m.wd, flags)
	if err != nil {
		return err
	}
	m.corePkg = pkg
	m.Pkgs = append(m.Pkgs, pkg)
	for {
		nextRoundDir := []string{}
		for _, d := range dirs {
			pkg, newDirs, err := m.parseDir(d, flags)
			if err != nil {
				return err
			}
			m.Pkgs = append(m.Pkgs, pkg)
			nextRoundDir = append(nextRoundDir, newDirs...)
		}
		if len(nextRoundDir) == 0 {
			break
		}
		dirs = nextRoundDir
	}
	return
}

func (m *Matte) importPathForDirectory(dir string) string {
	if dir == "./" {
		return m.modFile.Module.Mod.Path
	}
	return filepath.Join(m.modFile.Module.Mod.Path, dir)
}

func (m *Matte) processProject() error {
	for _, pkg := range m.Pkgs {
		m.currentPkg = pkg
		m.imports += fmt.Sprintf(`"%v"`, pkg.ImportPath) + "\n"
		for _, file := range pkg.Files {
			err := m.processFile(file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// extension of parser.parseDir which returns the additional directories inside the project
// also ignores _test.go files
// also parses go.mod file if exists, if a pkg has mod file it'll be attached to Pkg
func (m *Matte) parseDir(dirPath string, mode parser.Mode) (pkg *Pkg, dirs []string, first error) {
	fset := m.fileSet
	pkg = &Pkg{
		Package: &ast.Package{
			Files: make(map[string]*ast.File),
		},
	}
	list, err := os.ReadDir(dirPath)
	dirs = make([]string, 0)
	if err != nil {
		return nil, nil, err
	}
	for _, d := range list {
		filePath := filepath.Join(dirPath, d.Name())
		if d.IsDir() && !strings.HasSuffix(d.Name(), MatteDir) {
			dirs = append(dirs, filepath.Join(dirPath, d.Name()))
			continue
		}
		if strings.HasSuffix(filePath, "_test.go") || !strings.HasSuffix(filePath, ".go") {
			continue
		}
		if src, err := parser.ParseFile(fset, filePath, nil, mode); err == nil {
			name := src.Name.Name
			importPath := m.importPathForDirectory(dirPath)
			pkg = &Pkg{
				Name:       name,
				ImportPath: importPath,
				Package: &ast.Package{
					Name:  name,
					Files: make(map[string]*ast.File),
				},
			}
			pkg.Files[filePath] = src
		} else if first == nil {
			first = err
		}
	}

	return
}

// processes a ast.File and finds each REST handler specific to the passed framework(ex:gin) and
// parses the swag comments, based on these comments mounts the handler in the framework automatically so you dont have to manually
func (m *Matte) processFile(astFile *ast.File) (err error) {
	for _, fnDecl := range astFile.Decls {
		// iterate over all the functions in the package but not methods

		if fnDecl, ok := fnDecl.(*ast.FuncDecl); ok && fnDecl.Recv == nil {
			if fnDecl.Doc != nil {
				err := m.ProcessFn(fnDecl)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}

func (m *Matte) DeferCleanUp() error {
	return os.RemoveAll(m.matteDir)
}

type Decorator struct {
	name string
	args []string
}

func ParseDecorator(s string) (decorator *Decorator, err error) {
	expr, err := parser.ParseExpr(s)
	if err != nil {
		return nil, err
	}
	expr2, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil, fmt.Errorf("invalid decorator")
	}
	nameIdent, ok := expr2.Fun.(*ast.Ident)
	if !ok {
		return nil, fmt.Errorf("invalid decorator: name doesnt match any available decorator")
	}
	decorator = &Decorator{name: nameIdent.Name, args: []string{}}
	for _, arg := range expr2.Args {
		lit, ok := arg.(*ast.BasicLit)
		if !ok {
			return nil, fmt.Errorf("invalid decorator: arg %v is not basic lit", arg)
		}
		decorator.args = append(decorator.args, lit.Value)
	}
	return decorator, nil
}

func ParseComment(com *ast.CommentGroup) (decorators map[string]*Decorator, err error) {
	decorators = map[string]*Decorator{}
	for _, c := range com.List {
		splitLine := strings.Split(c.Text, "@")
		if len(splitLine) == 1 {
			// not a comment which we should process
			return
		}
		// ignore the first element
		splitLine = splitLine[1:]

		for _, possibleDecoratorText := range splitLine {
			var decorator *Decorator
			decorator, err = ParseDecorator(possibleDecoratorText)
			if err != nil {
				return
			}
			decorators[decorator.name] = decorator
		}
	}
	return
}

func (m *Matte) ProcessFn(fnDecl *ast.FuncDecl) (err error) {
	decorators := map[string]*Decorator{}
	if fnDecl.Doc != nil {
		decorators, err = ParseComment(fnDecl.Doc)
		if err != nil {
			return
		}
		if len(decorators) == 0 {
			// not a handler
			return
		}
		pathDecorator := decorators["path"]
		if pathDecorator == nil {
			err = fmt.Errorf("a 'path' decorator is required")
			return
		}
		err := m.ProcessPath(pathDecorator, fnDecl)
		if err != nil {
			return err
		}
	}
	return
}
