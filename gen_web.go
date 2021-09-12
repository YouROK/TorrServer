package main

import (
	"fmt"
	"io/fs"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func main() {
	dir, _ := os.Getwd()
	os.Chdir("web")
	if run("yarn") != nil {
		os.Exit(1)
	}
	if run("yarn", "run", "build") != nil {
		os.Exit(1)
	}
	os.Chdir(dir)

	compileHtml := "web/build/"
	srcGo := "server/web/pages/"

	run("rm", "-rf", srcGo+"template/pages")
	run("cp", "-r", compileHtml, srcGo+"template/pages")
	run("cp", "-r", "server/web/msx/pages", srcGo+"template/pages/msx")

	files := make([]string, 0)

	filepath.WalkDir(srcGo+"template/pages/", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			name := strings.TrimPrefix(path, srcGo+"template/")
			if !strings.HasPrefix(name, ".") {
				files = append(files, name)
			}
		}
		return nil
	})
	fmap := writeEmbed(srcGo+"template/html.go", files)
	writeRoute(srcGo+"template/route.go", fmap)
}

func writeEmbed(fname string, files []string) map[string]string {
	ff, err := os.Create(fname)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer ff.Close()
	embedStr := `package template

import (
	_ "embed"
)
`
	ret := make(map[string]string)

	for _, f := range files {
		fname := cleanName(strings.TrimPrefix(f, "pages"))
		embedStr += "\n//go:embed " + f + "\nvar " + fname + " []byte\n"
		ret[strings.TrimPrefix(f, "pages")] = fname
	}

	ff.WriteString(embedStr)
	return ret
}

func writeRoute(fname string, fmap map[string]string) {
	ff, err := os.Create(fname)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer ff.Close()
	embedStr := `package template

import (
	"github.com/gin-gonic/gin"
)

func RouteWebPages(route *gin.RouterGroup) {
	route.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})
`
	mime.AddExtensionType(".map", "application/json")
	mime.AddExtensionType(".webmanifest", "application/manifest+json")
	// sort fmap
	keys := make([]string, 0, len(fmap))
	for key := range fmap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, link := range keys {
		fmime := mime.TypeByExtension(filepath.Ext(link))
		embedStr += `
	route.GET("` + link + `", func(c *gin.Context) {
		c.Data(200, "` + fmime + `", ` + fmap[link] + `)
	})
`
	}
	embedStr += "}\n"

	ff.WriteString(embedStr)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func cleanName(fn string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return strings.Title(reg.ReplaceAllString(fn, ""))
}
