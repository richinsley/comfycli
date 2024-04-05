package env

import (
	_ "embed"

	"github.com/richinsley/comfycli/pkg"
)

var (
	CLIOptions *pkg.ComfyOptions
)

func SetLocalOptions(options *pkg.ComfyOptions) {
	CLIOptions = options
}
