package pkg

import (
	"bufio"
	"os"

	"github.com/richinsley/comfy2go/client"
	"github.com/spf13/viper"
)

type ComfyOptions struct {
	Host           string
	Port           int
	Json           bool
	PrettyJson     bool
	API            string
	APIValues      string
	GraphOutPath   string
	InlineImages   bool
	NoSaveData     bool
	DataToStdout   bool
	HomePath       string
	RecipesPath    string
	Yes            bool // Automatically answer yes on prompted questions
	OutputNodes    string
	NoSharedModels bool
	// API sub command options
	APIValuesOnly bool // only output the values of the API nodes
	Stdin         *bufio.Reader
	Client        *client.ComfyClient
}

func (o *ComfyOptions) ApplyEnvironment() {
	// Check if viper has a value for each setting, if so, use it to set the struct's fields
	if viper.IsSet("host") {
		o.Host = viper.GetString("host")
	}
	if viper.IsSet("port") {
		o.Port = viper.GetInt("port")
	}
	if viper.IsSet("pretty") {
		o.PrettyJson = viper.GetBool("pretty")
	}
	if viper.IsSet("api") {
		o.API = viper.GetString("api")
	}
	if viper.IsSet("graphout") {
		o.GraphOutPath = viper.GetString("graphout")
	}
	if viper.IsSet("inlineimages") {
		o.InlineImages = viper.GetBool("inlineimages")
	}
	if viper.IsSet("nosavedata") {
		o.NoSaveData = viper.GetBool("nosavedata")
	}
}

func (o *ComfyOptions) GetStdinReader() *bufio.Reader {
	if o.Stdin != nil {
		return o.Stdin
	}

	o.Stdin = bufio.NewReader(os.Stdin)
	return o.Stdin
}
