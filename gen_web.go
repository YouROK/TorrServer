package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	dir, _ := os.Getwd()

	if len(os.Args) > 0 {
		// Here you can be more cunning, but it will work anyway, for a clean build you need to clean the build folder using the --clean command
		if slices.ContainsFunc(os.Args, func(s string) bool {
			return s == "--clean" || s == "-c"
		}) {
			// There are problems with running under windows
			if err := run("rm", "-rf", "web/build"); err != nil {
				if strings.Contains(err.Error(), "executable file not found") {
					// Adding the ability to run on Windows with standard Go commands
					if err := os.RemoveAll("web/build"); err != nil {
						log.Default().Fatalln(err.Error())
					}
				} else {
					log.Default().Fatalln(err.Error())
				}
			}
		}
	}

	if _, err := os.Stat("web/build/static"); os.IsNotExist(err) {
		os.Chdir("web")
		if err = run("yarn"); err != nil {
			log.Default().Fatalln(err.Error())
		}
		if err = run("yarn", "run", "build"); err != nil {
			log.Default().Fatalln(err.Error())
		}
		os.Chdir(dir)
	}

	compileHtml := "web/build/"
	srcGo := "server/web/pages/"

	// There are problems with running under windows
	if err := run("rm", "-rf", srcGo+"template/pages"); err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			// Adding the ability to run on Windows with standard Go commands
			if err = os.RemoveAll(srcGo + "template/pages"); err != nil {
				log.Default().Fatalln(err.Error())
			}
		} else {
			log.Default().Fatalln(err.Error())
		}
	}
	// There are problems with running under windows
	if err := run("cp", "-r", compileHtml, srcGo+"template/pages/"); err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			// Adding the ability to run on Windows with standard Go commands
			if err = os.CopyFS(srcGo+"template/pages/", os.DirFS(filepath.Dir(compileHtml))); err != nil {
				log.Default().Fatalln(err.Error())
			}
		} else {
			log.Default().Fatalln(err.Error())
		}
	}
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
