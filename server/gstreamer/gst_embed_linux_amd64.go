//go:build linux && amd64 && embed_gstlib

package gstreamer

import _ "embed"

//go:embed embedded_gstlib_linux_amd64.zip
var embeddedGSTLibZip []byte
