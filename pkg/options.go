package pkg

import (
	"bufio"
	"os"
	"sync"

	"github.com/richinsley/comfy2go/client"
	"github.com/spf13/viper"
)

type ComfyOptions struct {
	Host           []string
	Port           []int
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
	RecipesRepos   string
	Yes            bool // Automatically answer yes on prompted questions
	GetVersion     bool
	OutputNodes    string
	NoSharedModels bool
	// path to a file to read from stdin
	StdinFile string
	// API sub command options
	APIValuesOnly    bool // only output the values of the API nodes
	Stdin            *bufio.Reader
	Clients          []*client.ComfyClient
	JsonScanner      *bufio.Scanner
	JsonScannerMutex *sync.Mutex
}

func (o *ComfyOptions) ApplyEnvironment() {
	// Check if viper has a value for each setting, if so, use it to set the struct's fields
	if viper.IsSet("host") {
		o.Host = viper.GetStringSlice("host")
		// o.Host = viper.GetString("host")
	}
	// if viper.IsSet("port") {
	// 	o.Port = viper.GetInt("port")
	// }
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

func (o *ComfyOptions) SetStdinReader(r *bufio.Reader) {
	o.Stdin = r
}

func (o *ComfyOptions) GetStdinReader() *bufio.Reader {
	if o.Stdin != nil {
		return o.Stdin
	}

	// if StdinFile is set, open the file and return the reader
	if o.StdinFile != "" {
		f, err := os.Open(o.StdinFile)
		if err != nil {
			return nil
		}
		o.Stdin = bufio.NewReader(f)
		return o.Stdin
	}

	// if no file is set, return the default stdin reader
	o.Stdin = bufio.NewReader(os.Stdin)
	return o.Stdin
}
