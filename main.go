package main

import (
	"fmt"
	"go/token"
	"log"
	"os"
	"time"

	m "github.com/ondbyte/matte/v1"
	flag "github.com/ondbyte/turbo_flag"
)

var matte *m.Matte

func main() {
	usage := `
matte: a microservice developement tooling for go
available sub commands are.
(run <sub-command> -h for more on it)
1. build`
	flag.MainCmd("matte", usage, flag.PanicOnError, os.Args[1:], matteCmd)

}

func matteCmd(cmd flag.CMD, args []string) {
	cmd.SubCmd("configure", `intialize your configuration for your app, this adds a config.matte.go to you root project, 
	where you can configure different frameworks and others configs`, configureCmd)
	cmd.SubCmd("build", "build your matte project", buildCmd)
	cmd.SubCmd("chinmaya", "wife: will something happen when i enter my name? can you make it work?", chinmayaCmd)
	err := cmd.Parse(args)
	if err != nil {
		panic(err)
	}
}

func configureCmd(cmd flag.CMD, args []string) {
	wd, err := os.Getwd()
	if err != nil {
		panic("unable to get working directory")
	}
	err = matte.Configure(wd)
	if err != nil {
		panic(err)
	}
}

func buildCmd(cmd flag.CMD, args []string) {
	help := false
	noBuild := false
	workingDir := ""
	cmd.BoolVar(&help, "help", false, "prints this", flag.Alias("h"))
	cmd.BoolVar(&noBuild, "no-build", false, "only runs finalize, this is a dev flag, possible to inspect src outputted in app.go ", flag.Alias("n"))
	cmd.StringVar(&workingDir, "dir", "./", "root directory of the project you need to build", flag.Alias("d"))
	err := cmd.Parse(args)
	if err != nil {
		panic(err)
	}
	if help {
		log.Println(cmd.GetDefaultUsage())
		return
	}

	fileset := token.NewFileSet()
	if err != nil {
		fmt.Println("unable to get working dir")
		panic(err)
	}
	err = os.Chdir(workingDir)
	if err != nil {
		panic(fmt.Errorf("unable change working dir to %v", workingDir))
	}
	matte, err = m.Build(fileset, frameworks.Frameworks{
		frameworks.NewHttp(),
	})
	if err != nil {
		panic(err)
	}

	err = matte.LoadProject()
	if err != nil {
		panic(err)
	}
	err = matte.ParseGeneralAPIInfo()
	if err != nil {
		panic(err)
	}

	err = matte.ProcessProject()
	if err != nil {
		panic(err)
	}

	err = matte.Finalize()
	if err != nil {
		panic(err)
	}
	if !noBuild {
		err = matte.Build()
		if err != nil {
			panic(err)
		}
	}

}

func chinmayaCmd(cmd flag.CMD, args []string) {
	for i := 0; i < 1000; i++ {
		fmt.Println("Yadu's wife")
		time.Sleep(time.Millisecond * 500)
	}
}
