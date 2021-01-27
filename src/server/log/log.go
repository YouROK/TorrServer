package log

import (
	"log"
	"os"
)

func Init(path string) {
	if path != "" {
		ff, err := os.Create(path)
		if err != nil {
			TLogln("Error create log file:", err)
			return
		}

		os.Stdout = ff
		os.Stderr = ff
		log.SetOutput(ff)
	}
}

func TLogln(v ...interface{}) {
	log.Println(v...)
}
