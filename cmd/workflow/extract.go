/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package workflow

import (
	"fmt"
	"os"
	"strings"

	"github.com/richinsley/comfy2go/client"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

func validatePNGArg(arg string) error {
	// Check the file extension
	if !strings.HasSuffix(arg, ".png") {
		return fmt.Errorf("the file must be a PNG file with a .png extension")
	}
	return nil
}

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract [png file path]",
	Short: "Extract a workflow from PNG metadata",
	Long: `Extract a workflow from PNG metadata.

example:
comfycli workflow extract /path/to/workflow.png > workflow.json`,
	// validate that a png file is provided
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("requires exactly one argument")
		}
		return validatePNGArg(args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		workflowPath := args[0]

		file, err := os.Open(workflowPath)
		if err != nil {
			slog.Error("Error reading PNG file", err)
			os.Exit(1)
		}
		defer file.Close()

		metadata, err := client.GetPngMetadata(file)
		if err != nil {
			slog.Error("Failed to extract metadat from PNG", err)
			os.Exit(1)
		}

		// validate that the metadata is a Comfy workflow by checking for a version key
		if _, ok := metadata["version"]; !ok {
			slog.Error("The provided PNG file does not contain Comfy workflow metadata")
			os.Exit(1)
		}

		fmt.Println(metadata)
	},
}

func InitExtract(workflowCmd *cobra.Command) {
	workflowCmd.AddCommand(extractCmd)
}
