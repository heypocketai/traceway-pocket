//go:build !localdist

package static

import (
	"embed"
	"io/fs"
)

//go:embed all:frontend
var staticFiles embed.FS

func GetStaticFS() (fs.FS, error) {
	return fs.Sub(staticFiles, "frontend")
}
