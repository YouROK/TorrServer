package template

import "embed"

//go:embed pages
var pages embed.FS

// Dlnaicon48png and Dlnaicon120png are used by the DLNA server.
var (
	Dlnaicon48png  []byte
	Dlnaicon120png []byte
)

func init() {
	Dlnaicon48png, _ = pages.ReadFile("pages/dlnaicon-48.png")
	Dlnaicon120png, _ = pages.ReadFile("pages/dlnaicon-120.png")
}
