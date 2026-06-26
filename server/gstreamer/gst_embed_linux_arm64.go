//go:build linux && arm64 && embed_gstlib

package gstreamer

import _ "embed"

//go:embed embedded_gstlib_linux_arm64.zip
var embeddedGSTLibZip []byte
