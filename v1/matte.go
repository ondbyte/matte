package matte

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/ondbyte/matte/frameworks"
	"github.com/ondbyte/matte/swaggo"
	"github.com/rogpeppe/go-internal/modfile"
	"github.com/swaggo/swag"
)

type Pkg struct {
	*ast.Package
	ImportPath string
	Name       string
}

type Matte struct {
	matteDir      string
	swagger       *spec.Swagger
	fileSet       *token.FileSet
	wd            string
	allFrameworks frameworks.Frameworks
	corePkg       *Pkg
	Pkgs          []*Pkg
	modFile       *modfile.File
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

func NewMatte(fileSet *token.FileSet, allFrameworks frameworks.Frameworks) (*Matte, error) {
	projectDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("unable get working dir due to err: %v", err)
	}
	matteDir, err := createMatteDir(projectDir)
	if err != nil {
		return nil, fmt.Errorf("unable to make matteDir due to err: %v", err)
	}
	return &Matte{
		fileSet:       fileSet,
		wd:            projectDir,
		allFrameworks: allFrameworks,
		matteDir:      matteDir,
	}, nil
}

// loads the project to matte's ast tree
// no test files will be included
func (m *Matte) LoadProject() (first error) {
	err := m.parseModFile()
	if err != nil {
		return err
	}
	flags := parser.AllErrors | parser.ParseComments
	m.Pkgs = []*Pkg{}
	pkg, dirs, err := m.parseDir("./", flags)
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

func (m *Matte) ProcessProject() error {
	for _, pkg := range m.Pkgs {
		for _, file := range pkg.Files {
			err := m.processFile(file, pkg.ImportPath, pkg.Name)
			if err != nil {
				return err
			}
		}
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
func (m *Matte) processFile(astFile *ast.File, pkgImportPathOfAstFile, pkgNameOfAstFile string) error {
	for _, fnDecl := range astFile.Decls {
		// iterate over all the functions in the package but not methods
		if fnDecl, ok := fnDecl.(*ast.FuncDecl); ok && fnDecl.Recv == nil {
			framework := m.allFrameworks.GetFrameworkForHandler(fnDecl.Type)
			if framework == nil {
				continue
			}
			if fnDecl.Doc == nil {
				return fmt.Errorf("handler %v does not have any comment, unable to parse swag spec for it", fnDecl.Name.Name)
			}
			operation := swag.NewOperation(nil)
			for _, c := range fnDecl.Doc.List {
				err := swaggo.ParseComment(operation, c.Text, astFile, "")
				if err != nil {
					return err
				}
			}
			for _, rp := range operation.RouterProperties {
				err := framework.AddHandler(rp.HTTPMethod, rp.Path, pkgImportPathOfAstFile, pkgNameOfAstFile, fnDecl.Name.Name)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}

// parses Matte function
func (m *Matte) ParseGeneralAPIInfo() error {
	generalApiInfoContainingFile := ""
	for _, file := range m.corePkg.Files {
		for _, decl := range file.Decls {
			if decl, ok := decl.(*ast.FuncDecl); ok && decl.Name.Name == GeneralApiInfoFuncName {
				generalApiInfoContainingFile = m.fileSet.Position(file.Package).Filename
			}
		}
	}
	if generalApiInfoContainingFile == "" {
		return fmt.Errorf("unable to find function name '%v', which is required to parse your applications core configurations like port it runs", GeneralApiInfoFuncName)
	}
	swaggoParser := swag.New()
	err := swaggoParser.ParseGeneralAPIInfo(generalApiInfoContainingFile)
	if err != nil {
		return fmt.Errorf("error while calling ParseGeneralAPIInfo : %v", err)
	}
	m.swagger = swaggoParser.GetSwagger()
	if m.swagger == nil {
		return fmt.Errorf("m.swagger should not be nil")
	}
	return nil
}

// finalizes the apps source code
// and writes to a matte.go file
func (m *Matte) Finalize() error {
	imports := []string{}
	funcNames := []string{}
	body := ""
	for _, v := range m.allFrameworks {
		importStmt, funcName, funcSrc, err := v.Finalize()
		if err != nil {
			return fmt.Errorf("unable to Finalize because err: %v", err)
		}
		imports = append(imports, importStmt)
		funcNames = append(funcNames, funcName)
		body += funcSrc + "\n"
	}
	src := "package main\n"
	for _, v := range imports {
		src += v + "\n"
	}
	src += "\n" + body + "\n\n"
	src += "func main(){\n"
	for _, v := range funcNames {
		src += "go " + v + "()\n"
	}
	src += "}\n"
	filePath, err := filepath.Abs(filepath.Join(m.matteDir, "app.go"))
	if err != nil {
		return fmt.Errorf("failed to get absolute path from path because err:%v", err)
	}
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %v because err: %v", filePath, err)
	}
	_, err = file.WriteString(src)
	if err != nil {
		return fmt.Errorf("failed to write to file %v because err: %v", filePath, err)
	}
	return nil
}

// builds the finalized main program written to matte.go using go build
func (m *Matte) Build() error {
	err := os.Chdir(m.matteDir)
	if err != nil {
		return fmt.Errorf("error while changing directory to matte dir %v, err: %v", m.matteDir, err)
	}
	defer os.Chdir(m.wd)
	build := exec.Command("go", "build")
	stdErr := &strings.Builder{}
	build.Stderr = stdErr
	err = build.Run()
	stdErrStr := stdErr.String()
	if err != nil {
		return fmt.Errorf("error running 'go build' command due to err: %w and stdErr: %v", err, stdErrStr)
	}
	return nil
}

//
func (m *Matte) DeferCleanUp() error {
	return os.RemoveAll(m.matteDir)
}
