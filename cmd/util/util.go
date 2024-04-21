package util

import (
	"github.com/richinsley/comfycli/pkg"
)

var (
	CLIOptions *pkg.ComfyOptions
)

func SetLocalOptions(options *pkg.ComfyOptions) {
	CLIOptions = options
}
