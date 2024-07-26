package embed_files

import (
	"embed"
)

//go:embed templates
var Templates embed.FS

//go:embed static
var Static embed.FS
