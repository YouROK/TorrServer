package main

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	dir, _ := os.Getwd()
	os.Chdir("web")
	run("npm", "run", "build-js")
	os.Chdir(dir)

	run("cp", "web/dest/index.html", "server/web/pages/template/pages/")
	// compileHtml := "web/dest/"
	// fs, _ := ioutil.ReadDir(compileHtml)
	// for _, f := range fs {
	// 	if strings.HasSuffix(f.Name(), ".html") {
	// 		name := filenameWithoutExtension(f.Name())
	// 		fmt.Println("Create template go:", "server/web/pages/template/"+name+"_html.go")
	// 		out, err := os.Create("server/web/pages/template/" + name + "_html.go")
	// 		if err != nil {
	// 			fmt.Println("Error create file", err)
	// 			os.Exit(1)
	// 			return
	// 		}
	//
	// 		fmt.Println("Read html:", compileHtml+f.Name())
	// 		buf, err := ioutil.ReadFile(compileHtml + f.Name())
	// 		if err != nil {
	// 			fmt.Println("Error read file", err)
	// 			os.Exit(1)
	// 			return
	// 		}
	// 		fmt.Println("Write template...")
	// 		out.Write([]byte("package template \n\nvar " + strings.Title(name) + "Html = []byte{"))
	// 		for _, b := range buf {
	// 			out.Write([]byte(strconv.Itoa(int(b)) + ", "))
	// 		}
	// 		out.Write([]byte("}"))
	// 		out.Close()
	//
	// 		fmt.Println("go fmt template...")
	// 		run("go", "fmt", "server/web/pages/template/"+name+"_html.go")
	// 		fmt.Println("Complete OK")
	// 	}
	// }
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func filenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(filepath.Base(fn), path.Ext(fn))
}
