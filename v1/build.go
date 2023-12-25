package matte

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func BuildProject(dir string, stdOut, stdErr io.Writer) error {
	/* cmd := exec.Command("go", "build", ".")
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr
	err = cmd.Run()
	if err != nil {
		return err
	} */
	return nil
}

// returns the package path of the a project at a dir
// this only returns if the dir is a golang project ie: has a go.mod file
func GetPackagePathFromGoMod(dir string) (string, error) {
	modFilePath := filepath.Join(dir, "go.mod")
	modFileBytes, err := ioutil.ReadFile(modFilePath)
	if err != nil {
		return "", fmt.Errorf("not a go project, unable to read go mod file due to err: %v", err)
	}
	importPath := modfile.ModulePath(modFileBytes)
	if importPath == "" {
		return "", fmt.Errorf("not a go project, unable to read go mod file due to err: %v", "no module path found")
	}
	return importPath, nil
}
