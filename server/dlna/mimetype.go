package dlna

import (
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

)

func init() {

// Add a minimal number of mime types to augment go's built in types
// for environments which don't have access to a mime.types file (e.g.
// Termux on android)
	for _, t := range []struct {
		mimeType   string
		extensions string
	}{
		{"audio/aac", ".aac"},
		{"audio/flac", ".flac"},
		{"audio/mpeg", ".mpga,.mpega,.mp2,.mp3,.m4a"},
		{"audio/ogg", ".oga,.ogg,.opus,.spx"},
		{"audio/opus", ".opus"},
		{"audio/weba", ".weba"},
		{"audio/x-wav", ".wav"},
		{"image/bmp", ".bmp"},
		{"image/gif", ".gif"},
		{"image/jpeg", ".jpg,.jpeg"},
		{"image/png", ".png"},
		{"image/tiff", ".tiff,.tif"},
		{"video/dv", ".dif,.dv"},
		{"video/fli", ".fli"},
		{"video/mpeg", ".mpeg,.mpg,.mpe"},
		{"video/mp2t", ".ts,.m2ts,.mts"},
		{"video/mp4", ".mp4"},
		{"video/quicktime", ".qt,.mov"},
		{"video/ogg", ".ogv"},
		{"video/webm", ".webm"},
		{"video/x-msvideo", ".avi"},
		{"video/x-matroska", ".mpv,.mkv"},
		{"text/srt", ".srt"},
	} {
		for _, ext := range strings.Split(t.extensions, ",") {
			err := mime.AddExtensionType(ext, t.mimeType)
			if err != nil {
				panic(err)
			}
		}
	}
	if err := mime.AddExtensionType(".rmvb", "application/vnd.rn-realmedia-vbr"); err != nil {
		log.Printf("Could not register application/vnd.rn-realmedia-vbr MIME type: %s", err)
	}
}

// Example: "video/mpeg"
type mimeType string

// IsMedia returns true for media MIME-types
func (mt mimeType) IsMedia() bool {
	return mt.IsVideo() || mt.IsAudio() || mt.IsImage()
}

// IsVideo returns true for video MIME-types
func (mt mimeType) IsVideo() bool {
	return strings.HasPrefix(string(mt), "video/") || mt == "application/vnd.rn-realmedia-vbr"
}

// IsAudio returns true for audio MIME-types
func (mt mimeType) IsAudio() bool {
	return strings.HasPrefix(string(mt), "audio/")
}

// IsImage returns true for image MIME-types
func (mt mimeType) IsImage() bool {
	return strings.HasPrefix(string(mt), "image/")
}

// Returns the group "type", the part before the '/'.
func (mt mimeType) Type() string {
	return strings.SplitN(string(mt), "/", 2)[0]
}

// Returns the string representation of this MIME-type
func (mt mimeType) String() string {
	return string(mt)
}

// MimeTypeByPath determines the MIME-type of file at the given path
func MimeTypeByPath(filePath string) (ret mimeType, err error) {
	ret = mimeTypeByBaseName(path.Base(filePath))
	if ret == "" {
		ret, err = mimeTypeByContent(filePath)
	}
	if ret == "video/x-msvideo" {
		ret = "video/avi"
	} else if ret == "" {
		ret = "application/octet-stream"
	}
	return
}

// Guess MIME-type from the extension, ignoring ".part".
func mimeTypeByBaseName(name string) mimeType {
	name = strings.TrimSuffix(name, ".part")
	ext := path.Ext(name)
	if ext != "" {
		return mimeType(mime.TypeByExtension(ext))
	}
	return mimeType("")
}

// Guess the MIME-type by analysing the first 512 bytes of the file.
func mimeTypeByContent(path string) (ret mimeType, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	var data [512]byte
	if n, err := file.Read(data[:]); err == nil {
		ret = mimeType(http.DetectContentType(data[:n]))
	}
	return
}
