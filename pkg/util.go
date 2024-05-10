package pkg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"bufio"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5"
	sixel "github.com/mattn/go-sixel"
	"github.com/richinsley/comfy2go/client"
	"github.com/richinsley/comfy2go/graphapi"
	kinda "github.com/richinsley/kinda/pkg"
	"github.com/schollz/progressbar/v3"
)

func ToJson(obj interface{}, purty bool) (string, error) {
	if purty {
		val, err := json.MarshalIndent(obj, "", "    ")
		if err != nil {
			return "", err
		}
		return string(val), nil
	} else {
		val, err := json.Marshal(obj)
		if err != nil {
			return "", err
		}
		return string(val), nil
	}
}

type CLIParameter struct {
	NodeID    int    // -1 for unset
	NodeTitle string // empty for unset
	API       bool   // is true if the parameter is an API parameter
	Name      string // the name of the parameter
	Value     string // the value of the parameter
}

func ParseParameters(params []string) []CLIParameter {
	var parsedParams []CLIParameter
	// Update the regex to match the full structure of parameters
	re := regexp.MustCompile(`^(?:(?:\((\d+)\))|([^:=]+):)?([^=]+)=(.+)$`)

	for _, param := range params {
		matches := re.FindStringSubmatch(param)
		if matches != nil {
			var cliParam CLIParameter
			if matches[1] != "" { // NodeID present
				nodeID, _ := strconv.Atoi(matches[1])
				cliParam = CLIParameter{
					NodeID: nodeID,
					API:    false,
				}
			} else if matches[2] != "" { // NodeTitle present
				cliParam = CLIParameter{
					NodeTitle: matches[2],
					API:       false,
					NodeID:    -1,
				}
			} else { // API parameter
				cliParam = CLIParameter{
					API:    true,
					NodeID: -1,
				}
			}
			cliParam.Name = matches[3]
			cliParam.Value = matches[4]
			parsedParams = append(parsedParams, cliParam)
		}
	}
	return parsedParams
}

func SaveData(data *[]byte, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	f.Write(*data)
	f.Close()
	return nil
}

func LoadData(path string) (*[]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func outputSixelImageToStd(data *[]byte) {
	// create an io.reader from the data bytes
	r := bytes.NewReader(*data)
	img, _, _ := image.Decode(r)
	sixel.NewEncoder(os.Stdout).Encode(img)
}

func outputInlineImageToStd(data *[]byte, name string, width int, height int) string {
	encoded_data := base64.StdEncoding.EncodeToString(*data)
	encoded_name := base64.StdEncoding.EncodeToString([]byte(name))
	sizestr := fmt.Sprintf("size=%d;", len(*data))
	namestr := fmt.Sprintf("name=%s;", encoded_name)
	dimstr := ""
	if width > 0 && height > 0 {
		dimstr = fmt.Sprintf("width=%dpx;height=%dpx;", width, height)
	}
	retv := fmt.Sprintf("\033]1337;File=inline=1;%s%s%spreserveAspectRatio=1:%s\a\n", sizestr, namestr, dimstr, encoded_data)

	return retv
}

func OutputInlineToStd(data *[]byte, name string, width int, height int) {
	os.Stdout.WriteString(outputInlineImageToStd(data, name, width, height))
}

type Workflow struct {
	ClientIndex int
	Client      *client.ComfyClient
	Graph       *graphapi.Graph
	SimpleAPI   *graphapi.SimpleAPI
}

func GetFullWorkflow(client_index int, options *ComfyOptions, workflow string, cb *client.ComfyClientCallbacks) (*Workflow, *[]string, error) {
	clientaddr := options.Host[client_index]
	clientport := options.Port[client_index]

	if options.Clients == nil {
		options.Clients = make([]*client.ComfyClient, len(options.Host))
	}

	// create a client if there is not one in options already
	var c *client.ComfyClient = nil
	if options.Clients[client_index] != nil {
		// resuse the client in options
		c = options.Clients[client_index]
	} else {
		// create a new client
		c = client.NewComfyClient(clientaddr, clientport, cb)
		options.Clients[client_index] = c
	}

	// the client needs to be in an initialized state before usage
	if !c.IsInitialized() {
		err := c.Init()
		if err != nil {
			return nil, nil, err
		}
	}

	// load the workflow
	var g *graphapi.Graph = nil
	var missing *[]string = nil
	var err error = nil
	// is workflow a png file?
	if strings.HasSuffix(strings.ToLower(workflow), ".png") {
		g, missing, err = c.NewGraphFromPNGFile(workflow)
	} else {
		g, missing, err = c.NewGraphFromJsonFile(workflow)
	}

	if err != nil {
		return nil, missing, err
	}

	simple_api := g.GetSimpleAPI(&options.API)

	// return the client and the graph
	return &Workflow{
		ClientIndex: client_index,
		Client:      c,
		Graph:       g,
		SimpleAPI:   simple_api,
	}, missing, nil
}

func setPropertValue(client *client.ComfyClient, options *ComfyOptions, prop graphapi.Property, value interface{}) (bool, error) {
	var readFromPipe bool = false
	var err error = nil

	switch prop.TypeString() {
	// "INT"			an int64
	// "FLOAT"			a float64
	// "STRING"			a single line, or multiline string
	// "COMBO"			one of a given list of strings
	// "BOOLEAN"		a labeled bool value
	// "IMAGEUPLOAD"	image uploader
	// "CASCADE"		collection cascading style properties
	// "UNKNOWN"		everything else (unsettable)
	case "INT":
		readFromPipe, err = SetGenericPropertValue(client, options, prop, value)
	case "FLOAT":
		readFromPipe, err = SetGenericPropertValue(client, options, prop, value)
	case "STRING":
		readFromPipe, err = SetGenericPropertValue(client, options, prop, value)
	case "COMBO":
		readFromPipe, err = SetGenericPropertValue(client, options, prop, value)
	case "BOOLEAN":
		readFromPipe, err = SetGenericPropertValue(client, options, prop, value)
	case "IMAGEUPLOAD":
		// "choose file to upload" or "file"
		readFromPipe, err = SetfileUploadPropertyValue(client, options, prop, value)
	case "CASCADE":
		readFromPipe, err = SetGenericPropertValue(client, options, prop, value)
	case "UNKNOWN":
		readFromPipe, err = SetGenericPropertValue(client, options, prop, value)
	default:
		readFromPipe, err = SetGenericPropertValue(client, options, prop, value)
	}
	return readFromPipe, err
}

func TestParametersHasPipeLoop(options *ComfyOptions, parameters []CLIParameter) (bool, error) {
	retv := false
	pipedparamcount := 0
	for _, param := range parameters {
		if param.Value == "-" {
			pipedparamcount += 1
		}
	}
	if pipedparamcount > 1 {
		return true, fmt.Errorf("only one parameter can read from stdin, use SimpleAPI for more")
	} else if pipedparamcount == 1 {
		retv = true
	}

	if options.APIValues != "" {
		if retv {
			return true, fmt.Errorf("APIValues and parameters reading from stdin cannot be used together")
		}
		if options.APIValues != "-" {
			// set the stdin reader to the file
			f, err := os.Open(options.APIValues)
			if err != nil {
				return false, err
			}
			options.SetStdinReader(bufio.NewReader(f))
			retv = true
		} else {
			// set the stdin reader to os.Stdin
			options.SetStdinReader(bufio.NewReader(os.Stdin))
			retv = true
		}
	}
	return retv, nil
}

func ApplyParameters(client *client.ComfyClient, options *ComfyOptions, graph *graphapi.Graph, simple_api *graphapi.SimpleAPI, parameters []CLIParameter) (bool, error) {
	// if we encounter any read from stdin, we need to set hasPipeLoop to true
	hasPipeLoop := false

	// if APIValues is defined, load the values as a map[string]interface{} and apply those first
	if options.APIValues != "" {
		if simple_api == nil {
			return false, fmt.Errorf("apivalues specified but no SimpleAPI provided or found in the graph")
		}

		// if the APIValues is set, try to read the values from stdin or the file
		var apivalues map[string]interface{} = nil
		if options.APIValues != "" {
			// prevent concurrent access to the scanner
			options.JsonScannerMutex.Lock()
			jobj, scanner, err := ScanJsonFromReader(options.GetStdinReader(), options.JsonScanner)
			options.JsonScanner = scanner
			options.JsonScannerMutex.Unlock()

			if err != nil {
				return false, err
			}
			if jobj == nil {
				return false, fmt.Errorf("no JSON object found in the input")
			}
			hasPipeLoop = true
			apivalues = jobj.(map[string]interface{})
		}

		// if there are api values, apply them first
		for k, v := range apivalues {
			targetprop, ok := simple_api.Properties[k]
			if ok {
				var err error
				var pl bool
				pl, err = setPropertValue(client, options, targetprop, v)
				if err != nil {
					return false, err
				}
				hasPipeLoop = hasPipeLoop || pl
			} else {
				slog.Error(fmt.Sprintf("Property %s not found in the SimpleAPI", k))
			}
		}
	}

	// apply the parameters to the graph
	for _, param := range parameters {
		if param.API {
			if prop, okparam := simple_api.Properties[param.Name]; okparam {
				var err error
				hasPipeLoop, err = setPropertValue(client, options, prop, param.Value)
				if err != nil {
					return false, err
				}
			} else {
				slog.Error(fmt.Sprintf("Property %s not found in the SimpleAPI", param.Name))
			}
		} else {
			var node *graphapi.GraphNode = nil
			if param.NodeID != -1 {
				node = graph.GetNodeById(param.NodeID)
			} else {
				node = graph.GetFirstNodeWithTitle(param.NodeTitle)
			}
			if node == nil {
				return false, fmt.Errorf("node %v not found in graph", param.NodeTitle)
			}
			// get the property interface for param.Name
			prop := node.GetPropertyWithName(param.Name)
			if prop == nil {
				return false, fmt.Errorf("property %v not found in node %v", param.Name, param.NodeTitle)
			}

			var err error
			hasPipeLoop, err = setPropertValue(client, options, prop, param.Value)
			if err != nil {
				return false, err
			}
		}
	}
	return hasPipeLoop, nil
}

func ReadAndDeserializeJSON(dst interface{}, jsonInput string) error {
	decoder := json.NewDecoder(strings.NewReader(jsonInput))
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("error decoding JSON: %w", err)
	}
	return nil
}

func ScanJsonFromReader(r io.Reader, scanner *bufio.Scanner) (interface{}, *bufio.Scanner, error) {
	if scanner == nil {
		scanner = bufio.NewScanner(r)
	}
	var jsonBlock strings.Builder

	var jobj interface{}
	for scanner.Scan() {
		line := scanner.Text()
		jsonBlock.WriteString(line)

		// Attempt to decode the current block
		if err := ReadAndDeserializeJSON(&jobj, jsonBlock.String()); err == nil {
			// Successfully decoded, process the person
			jsonBlock.Reset() // Reset the block for the next JSON object
			break
		}
		// If decoding fails, it might be because the JSON object is split across lines, so continue accumulating
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading standard input: %v\n", err)
		return nil, scanner, err
	}

	return jobj, scanner, nil
}

// func ListFiles(path string, topOnly bool) ([]string, error) {
// 	var files []string
// 	if topOnly {
// 		// Read only the top level of the directory.
// 		entries, err := os.ReadDir(path)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for _, entry := range entries {
// 			if !entry.IsDir() {
// 				files = append(files, filepath.Join(path, entry.Name()))
// 			}
// 		}
// 	} else {
// 		// Walk through the directory recursively.
// 		err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
// 			if err != nil {
// 				return err
// 			}
// 			if !info.IsDir() {
// 				files = append(files, path)
// 			}
// 			return nil
// 		})
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return files, nil
// }

func ListFiles(path string, topOnly bool, relative bool) ([]string, error) {
	var files []string
	if topOnly {
		// Read only the top level of the directory.
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				// Ignore hidden files
				if relative {
					files = append(files, entry.Name())
				} else {
					files = append(files, filepath.Join(path, entry.Name()))
				}
			}
		}
	} else {
		// Walk through the directory recursively.
		err := filepath.Walk(path, func(currentPath string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Skip hidden files and directories
			if strings.HasPrefix(filepath.Base(currentPath), ".") {
				if info.IsDir() {
					return filepath.SkipDir // Skip the entire directory if it is hidden
				}
				return nil // Skip hidden files
			}
			if !info.IsDir() {
				if relative {
					relPath, err := filepath.Rel(path, currentPath)
					if err != nil {
						return err
					}
					files = append(files, relPath)
				} else {
					files = append(files, currentPath)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

// FindAvailableEphemeralPort finds an available ephemeral port on the local loopback device.
// It returns the port number if found, or an error otherwise.
func FindAvailableEphemeralPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("failed to listen on a port: %v", err)
	}
	defer listener.Close()

	_, portString, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		return 0, fmt.Errorf("failed to parse listener address: %v", err)
	}

	return net.LookupPort("tcp", portString)
}

// downloadFile downloads a file from the specified URL and saves it to the given target path.
// It also handles redirects with custom logic.
func DownloadFile(url string, targetPath string, max_redirects int, feedback kinda.CreateEnvironmentOptions) error {
	if feedback == kinda.ShowProgressBar || feedback == kinda.ShowProgressBarVerbose {
		_, file := filepath.Split(targetPath)
		// Custom HTTP client with redirect policy.
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= max_redirects { // Limit the number of redirects.
					return http.ErrUseLastResponse
				}
				if feedback == kinda.ShowVerbose || feedback == kinda.ShowProgressBarVerbose {
					fmt.Println("Redirected to:", req.URL)
				}
				return nil // Allow redirect.
			},
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error downloading file: %v", err)
		}
		defer resp.Body.Close()

		f, _ := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, 0644)
		defer f.Close()

		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			fmt.Sprintf("Downloading %s", file),
		)
		io.Copy(io.MultiWriter(f, bar), resp.Body)
	} else {
		// Custom HTTP client with redirect policy.
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= max_redirects { // Limit the number of redirects.
					return http.ErrUseLastResponse
				}
				if feedback == kinda.ShowVerbose || feedback == kinda.ShowProgressBarVerbose {
					fmt.Println("Redirected to:", req.URL)
				}
				return nil // Allow redirect.
			},
		}

		// Make a GET request using the custom client.
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check if the download was successful.
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned non-200 status: %d %s", resp.StatusCode, resp.Status)
		}

		// Create the file where the contents will be saved.
		outFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		// Copy the contents from the response body to the file.
		_, err = io.Copy(outFile, resp.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

// create a slice of strings that contains all the unique strings from s1 and s2
func UnionStringSlices(s1 []string, s2 []string) []string {
	m := make(map[string]bool)
	for _, s := range s1 {
		m[s] = true
	}
	for _, s := range s2 {
		m[s] = true
	}
	retv := make([]string, 0)
	for k := range m {
		retv = append(retv, k)
	}
	return retv
}

// YesNo - prompt the user for a yes or no response
func YesNo(prompt string, default_yes bool) (bool, error) {
	var default_str string
	if default_yes {
		default_str = "Y"
	} else {
		default_str = "N"
	}
	fmt.Printf("%s [%s]: ", prompt, default_str)
	var response string
	fmt.Scanf("%s", &response)
	// to upper
	response = strings.ToUpper(response)
	if response == "" {
		response = default_str
	}

	if response == "Y" {
		return true, nil
	} else if response == "N" {
		return false, nil
	} else {
		return false, fmt.Errorf("invalid response")
	}
}

// OneOf - given a list of options, prompt the user to select one, if default is -1 then
// there is no default selection
func OneOf(values []string, default_index int) (string, error) {
	if len(values) == 0 {
		return "", fmt.Errorf("no values to select from")
	}
	if default_index >= len(values) {
		default_index = len(values) - 1
	}
	for i, v := range values {
		fmt.Printf("%d: %s\n", i, v)
	}
	if default_index == -1 {
		fmt.Printf("Select one: ")
	} else {
		fmt.Printf("Select one [%d]: ", default_index)
	}

	var selection int
	_, err := fmt.Scanf("%d", &selection)
	if err != nil {
		if default_index == -1 {
			return "", fmt.Errorf("invalid selection")
		}
		selection = default_index
	}

	// check if the selection is within the range of values
	if selection < 0 || selection >= len(values) {
		return "", fmt.Errorf("invalid selection")
	}
	return values[selection], nil
}

func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func CloneRepo(repoURL, repoPath string, showoutput bool) (*git.Repository, error) {
	cloneoptions := &git.CloneOptions{
		URL: repoURL,
	}

	if showoutput {
		// output clone status to stdout
		cloneoptions.Progress = os.Stdout
	}

	repo, err := git.PlainClone(repoPath, false, cloneoptions)

	if err != nil {
		return nil, err
	}

	return repo, nil
}
