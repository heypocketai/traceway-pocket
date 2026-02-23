//go:build localdist

package static

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var staticFiles embed.FS

func GetStaticFS() (fs.FS, error) {
	return fs.Sub(staticFiles, "dist")
}
