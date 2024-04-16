package workflow

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"image/png"
	"io"
	"os"
	"strings"

	"github.com/richinsley/comfy2go/client"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

var pngoutfile string = ""

// AddPngMetadata takes an existing PNG reader and a map of metadata,
// and returns a new PNG file content with updated tEXt chunks.
func AddPngMetadata(r io.Reader, metadata map[string]string) ([]byte, error) {
	var buffer bytes.Buffer

	// Read and validate PNG header
	header := make([]byte, 8)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(header, []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		return nil, errors.New("not a valid PNG file")
	}
	buffer.Write(header)

	// Process existing chunks
	processedChunks := make(map[string]bool)
	for {
		var length uint32
		err := binary.Read(r, binary.BigEndian, &length)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		chunkType := make([]byte, 4)
		_, err = io.ReadFull(r, chunkType)
		if err != nil {
			return nil, err
		}

		chunkData := make([]byte, length)
		_, err = io.ReadFull(r, chunkData)
		if err != nil {
			return nil, err
		}

		// Read the CRC but do not use it, as it will be recalculated
		oldCRC := make([]byte, 4)
		_, err = io.ReadFull(r, oldCRC)
		if err != nil {
			return nil, err
		}

		if string(chunkType) == "tEXt" {
			keywordEnd := bytes.IndexByte(chunkData, 0)
			if keywordEnd != -1 {
				keyword := string(chunkData[:keywordEnd])
				if newContent, ok := metadata[keyword]; ok {
					// Update chunk with new metadata
					newChunkData := append([]byte(keyword), 0)
					newChunkData = append(newChunkData, []byte(newContent)...)
					writeChunk(&buffer, "tEXt", newChunkData)
					processedChunks[keyword] = true
					continue
				}
			}
		}

		// Write unmodified chunk
		writeChunk(&buffer, string(chunkType), chunkData)
	}

	// Add new tEXt chunks for any unprocessed metadata
	for key, value := range metadata {
		if _, processed := processedChunks[key]; !processed {
			newChunkData := append([]byte(key), 0)
			newChunkData = append(newChunkData, []byte(value)...)
			writeChunk(&buffer, "tEXt", newChunkData)
		}
	}

	return buffer.Bytes(), nil
}

func writeChunk(buffer *bytes.Buffer, chunkType string, chunkData []byte) {
	length := uint32(len(chunkData))
	binary.Write(buffer, binary.BigEndian, length)
	buffer.WriteString(chunkType)
	buffer.Write(chunkData)

	// Calculate and write the CRC
	crc := crc32.NewIEEE()
	crc.Write([]byte(chunkType))
	crc.Write(chunkData)
	crcSum := crc.Sum32()
	binary.Write(buffer, binary.BigEndian, crcSum)
}

// injectCmd represents the extract command
var injectCmd = &cobra.Command{
	Use:   "inject [workflow.json] [png file path]",
	Short: "Inject a workflow into PNG metadata",
	Long: `Inject a workflow into PNG metadata.
If no output file is specified with "--output", the new PNG file will be written to stdout.

examples:

# Load a workflow from a png file, inject it into a PNG file and write the new PNG to stdout
comfycli workflow inject workflow.png image.png > newimage.png

# Load a workflow from a json file, inject it into a PNG file and write the new PNG to newimage.png
comfycli workflow inject workflow.json image.png --output newimage.png
`,
	// validate that a png file is provided
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("requires two arguemnts [workflow.json/workflow.png] [png input file]")
		}
		return validatePNGArg(args[1])
	},
	Run: func(cmd *cobra.Command, args []string) {
		workflowPath := args[0]
		pngPath := args[1]

		// Read the workflow JSON
		// is the workflow path a json or png file
		var workflow map[string]string
		tolower := strings.ToLower(workflowPath)
		if strings.HasSuffix(tolower, ".json") {
			// Read the workflow JSON file into workflow using ioutil.ReadFile
			data, err := os.ReadFile(workflowPath)
			if err != nil {
				slog.Error("Error reading workflow JSON file", err)
				os.Exit(1)
			}

			// parse the workflow JSON file to ensure it is valid
			err = json.Unmarshal(data, &map[string]interface{}{})
			if err != nil {
				slog.Error("Error parsing workflow JSON file", err)
				os.Exit(1)
			}

			// we don't need tworkflow anymore, we'll just use the map[string]string version
			workflow = make(map[string]string)
			workflow["workflow"] = string(data)
			workflow["prompt"] = "{}"

		} else if strings.HasSuffix(tolower, ".png") {
			file, err := os.Open(workflowPath)
			if err != nil {
				slog.Error("Error reading PNG file", err)
				os.Exit(1)
			}
			defer file.Close()

			workflow, err = client.GetPngMetadata(file)
			if err != nil {
				slog.Error("Failed to extract metadat from PNG", "error", err)
				os.Exit(1)
			}
		}

		// decode the png file at pngPath into an image.Image then re-encode it to remove any existing metadata
		// and add the new metadata
		// Decode the file as a PNG.
		file, err := os.Open(pngPath)
		if err != nil {
			slog.Error(fmt.Sprintf("could not open the file: %v", err))
			os.Exit(1)
		}
		defer file.Close()
		img, err := png.Decode(file)
		if err != nil {
			slog.Error(fmt.Sprintf("could not decode the file as PNG: %v", err))
			os.Exit(1)
		}

		// re-encode the omage into a buffer
		var buffer bytes.Buffer
		err = png.Encode(&buffer, img)
		if err != nil {
			slog.Error(fmt.Sprintf("could not re-encode the image: %v", err))
			os.Exit(1)
		}

		newpng, err := AddPngMetadata(&buffer, workflow)
		if err != nil {
			slog.Error("Failed to add metadata to PNG", err)
			os.Exit(1)
		}

		// Write the new PNG file to stdout
		if pngoutfile != "" {
			err := os.WriteFile(pngoutfile, newpng, 0644)
			if err != nil {
				slog.Error("Failed to write new PNG file", err)
				os.Exit(1)
			}
		} else {
			os.Stdout.Write(newpng)
		}
	},
}

func InitInject(workflowCmd *cobra.Command) {
	workflowCmd.AddCommand(injectCmd)

	injectCmd.PersistentFlags().StringVarP(&pngoutfile, "output", "o", "", "Path to write new PNG file with metadata")
}
