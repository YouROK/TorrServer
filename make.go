package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Need version")
		os.Exit(1)
	}
	release_version := os.Args[1]
	isTest := false
	if len(os.Args) > 2 {
		isTest = true
	}

	fmt.Println("\nMake:", release_version, "\n")
	cmd := exec.Command("./build-all.sh")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("Error compile", err)
		os.Exit(1)
	}
	files, err := ioutil.ReadDir("dist")
	if err != nil {
		fmt.Println("Error read dist")
		os.Exit(1)
	}

	fmt.Println("\nMake json")
	js := Release{}
	js.Name = "TorrServer"
	js.Version = release_version
	js.BuildDate = time.Now().Format("02.01.2006")
	js.Links = make(map[string]string)
	for _, f := range files {
		arch := strings.TrimPrefix(f.Name(), "TorrServer-")
		js.Links[arch] = "https://github.com/YouROK/TorrServer/releases/download/" + release_version + "/" + f.Name()
	}
	buf, err := json.MarshalIndent(&js, "", " ")
	if err != nil {
		fmt.Println("Error make json")
		os.Exit(1)
	}
	if isTest {
		err = ioutil.WriteFile("test.json", buf, 0666)
	} else {
		err = ioutil.WriteFile("release.json", buf, 0666)
	}
	if err != nil {
		fmt.Println("Error write to json file:", err)
		os.Exit(1)
	}
	fmt.Println("\n\nEnter tag manually:\n")
	fmt.Println("git push origin", release_version)
	fmt.Println()
}

type Release struct {
	Name      string
	Version   string
	BuildDate string
	Links     map[string]string
}

//"update": {
//"name": "TorrServer",
//"version": "1.0.61",
//"build_date": "20.07.2018",
//"links":[
//{"android-386":"https://github.com/YouROK/TorrServe/releases/download/1.0.61/TorrServer-android-386"},
//{"android-amd64":"https://github.com/YouROK/TorrServe/releases/download/1.0.61/TorrServer-android-amd64"},
//{"android-arm7":"https://github.com/YouROK/TorrServe/releases/download/1.0.61/TorrServer-android-arm7"},
//{"android-arm64":"https://github.com/YouROK/TorrServe/releases/download/1.0.61/TorrServer-android-arm64"},
//]
//}
