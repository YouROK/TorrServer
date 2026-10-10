package main

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"log"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
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
		} else {
			// Do not uncomment, be aware - its crash the build
			// log.Default().Fatalln("Wrong args?")
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

	files := make([]string, 0)

	filepath.WalkDir(srcGo+"template/pages/", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			name := strings.TrimPrefix(path, srcGo+"template/")
			if strings.Contains(name, "\\") {
				// Adding the ability to run on Windows with standard Go commands
				name = strings.TrimPrefix(strings.ReplaceAll(name, "\\", "/"), "server/web/pages/template/")
			}
			if !strings.HasPrefix(filepath.Base(name), ".") {
				files = append(files, name)
			}
		}
		return nil
	})
	sort.Strings(files)
	fmap := writeEmbed(srcGo+"template/html.go", files)
	writeRoute(srcGo+"template/route.go", srcGo+"template/pages", fmap)
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

func writeRoute(fname, pagesDir string, fmap map[string]string) {
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

func RouteWebPages(route gin.IRouter) {
	route.GET("/", func(c *gin.Context) {
		serve(c, Indexhtml, ` + etagOf(pagesDir+"/index.html") + `, "text/html; charset=utf-8", false)
	})
`
	// Pin .js so the output does not depend on the Go version or the system MIME table
	mime.AddExtensionType(".js", "application/javascript")
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
		if fmime == "application/xml" || fmime == "application/javascript" {
			fmime = fmime + "; charset=utf-8"
		}
		if fmime == "image/x-icon" {
			fmime = "image/vnd.microsoft.icon"
		}
		// Only the build output under static/ has content hashes in its file names
		immutable := strings.HasPrefix(link, "/static/")
		embedStr += `
	route.GET("` + link + `", func(c *gin.Context) {
		serve(c, ` + fmap[link] + `, ` + etagOf(pagesDir+link) + `, "` + fmime + `", ` + fmt.Sprint(immutable) + `)
	})
`
	}
	embedStr += "}\n"

	ff.WriteString(embedStr)
}

// etagOf returns a quoted strong ETag for a file, as a Go string literal
func etagOf(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Default().Fatalln(err.Error())
	}
	return fmt.Sprintf("`\"%x\"`", md5.Sum(data))
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
