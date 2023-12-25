package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	os.Args = []string{"matte", "build", "--no-build", "--dir", "./test_project"}
	main()
}
