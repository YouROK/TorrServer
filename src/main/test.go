package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"server/utils"
	"server/version"

	"github.com/anacrolix/torrent"
)

/*
{
 "Name": "TorrServer",
 "Version": "1.1.71",
 "BuildDate": "17.05.2019",
 "Links": {
  "android-386": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-android-386",
  "android-amd64": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-android-amd64",
  "android-arm64": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-android-arm64",
  "android-arm7": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-android-arm7",
  "darwin-amd64": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-darwin-amd64",
  "linux-386": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-linux-386",
  "linux-amd64": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-linux-amd64",
  "linux-arm5": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-linux-arm5",
  "linux-arm6": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-linux-arm6",
  "linux-arm64": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-linux-arm64",
  "linux-arm7": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-linux-arm7",
  "linux-mips": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-linux-mips",
  "linux-mips64": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-linux-mips64",
  "linux-mips64le": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-linux-mips64le",
  "linux-mipsle": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-linux-mipsle",
  "windows-386.exe": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-windows-386.exe",
  "windows-amd64.exe": "https://github.com/YouROK/TorrServer/releases/download/1.1.71/TorrServer-windows-amd64.exe"
 }
}
*/

type release struct {
	Version string
	Links   map[string]string
}

func mkReleasesJS() {
	var releases []release
	for i := 65; i <= version.VerInt; i++ {
		links := map[string]string{
			"android-386":       fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-android-386", i),
			"android-amd64":     fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-android-amd64", i),
			"android-arm64":     fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-android-arm64", i),
			"android-arm7":      fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-android-arm7", i),
			"darwin-amd64":      fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-darwin-amd64", i),
			"linux-386":         fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-linux-386", i),
			"linux-amd64":       fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-linux-amd64", i),
			"linux-arm5":        fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-linux-arm5", i),
			"linux-arm6":        fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-linux-arm6", i),
			"linux-arm64":       fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-linux-arm64", i),
			"linux-arm7":        fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-linux-arm7", i),
			"linux-mips":        fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-linux-mips", i),
			"linux-mips64":      fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-linux-mips64", i),
			"linux-mips64le":    fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-linux-mips64le", i),
			"linux-mipsle":      fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-linux-mipsle", i),
			"windows-386.exe":   fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-windows-386.exe", i),
			"windows-amd64.exe": fmt.Sprintf("https://github.com/YouROK/TorrServer/releases/download/1.1.%d/TorrServer-windows-amd64.exe", i),
		}
		rel := release{
			Version: fmt.Sprintf("1.1.%d", i),
			Links:   links,
		}
		releases = append(releases, rel)
	}
	buf, _ := json.MarshalIndent(releases, "", " ")
	if len(buf) > 0 {
		ioutil.WriteFile("/home/yourok/surge/torrserve/releases.json", buf, 0666)
	}
}

func test() {
	config := torrent.NewDefaultClientConfig()

	config.EstablishedConnsPerTorrent = 100
	config.HalfOpenConnsPerTorrent = 65
	config.DisableIPv6 = true
	config.NoDHT = true

	client, err := torrent.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	//Ubuntu
	t, err := client.AddMagnet("magnet:?xt=urn:btih:e4be9e4db876e3e3179778b03e906297be5c8dbe&dn=ubuntu-18.04-desktop-amd64.iso&tr=http%3a%2f%2ftorrent.ubuntu.com%3a6969%2fannounce&tr=http%3a%2f%2fipv6.torrent.ubuntu.com%3a6969%2fannounce")
	if err != nil {
		log.Fatal(err)
	}
	<-t.GotInfo()
	file := t.Files()[0]

	reader := file.NewReader()
	var wa sync.WaitGroup
	wa.Add(1)

	go func() {
		buf := make([]byte, t.Info().PieceLength)
		for {
			_, err := reader.Read(buf)
			if err != nil {
				break
			}
		}
		wa.Done()
	}()

	go func() {
		cl := t.Closed()
		lastTimeSpeed := time.Now()
		DownloadSpeed := 0.0
		BytesReadUsefulData := int64(0)
		for {
			select {
			case <-cl:
				return
			default:
				client.WriteStatus(os.Stdout)
				st := t.Stats()
				deltaDlBytes := st.BytesReadUsefulData.Int64() - BytesReadUsefulData
				deltaTime := time.Since(lastTimeSpeed).Seconds()
				DownloadSpeed = float64(deltaDlBytes) / deltaTime
				BytesReadUsefulData = st.BytesReadUsefulData.Int64()
				lastTimeSpeed = time.Now()
				fmt.Println("DL speed:", utils.Format(DownloadSpeed))
			}
			time.Sleep(time.Second)
		}
	}()
	wa.Wait()
}
