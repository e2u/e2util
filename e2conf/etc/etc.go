package etc

import (
	"embed"
)

var (
	//go:embed *.toml
	Fs embed.FS
)
