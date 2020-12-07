package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	fs, _ := ioutil.ReadDir("web")
	for _, f := range fs {
		if strings.HasSuffix(f.Name(), ".html") {
			name := filenameWithoutExtension(f.Name())
			out, err := os.Create("src/server/web/pages/template/" + name + "_html.go")
			if err != nil {
				fmt.Println("Error create file", err)
				os.Exit(1)
				return
			}
			out.Write([]byte("package template \n\nvar " + strings.Title(name) + "Html = []byte{"))
			buf, err := ioutil.ReadFile("web/" + f.Name())
			if err != nil {
				fmt.Println("Error read file", err)
				os.Exit(1)
				return
			}
			for _, b := range buf {
				out.Write([]byte(strconv.Itoa(int(b)) + ", "))
			}
			out.Write([]byte("}"))
			out.Close()
		}
	}
}

func filenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(filepath.Base(fn), path.Ext(fn))
}
