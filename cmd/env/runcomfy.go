/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package env

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// runcomfyCmd
var runcomfyCmd = &cobra.Command{
	Use:   "runcomfy",
	Short: "Launch ComfyUI within an environment",
	Long: `Launch ComfyUI within an environment.
	You can pass additional arguments to the ComfyUI script by placing them after '--'

	examples:
	# run ComfyUI in the default environment
	comfycli env runcomfy --env default -- --help
	
	# get ComfyUI command line help
	comfycli env runcomfy --env default -- --help

	# run ComfyUI in the myenv environment.  Pass the --listen and --highvram arguments to the ComfyUI script
	comfycli env runcomfy myenv -- --listen --highvram`,
	Run: func(cmd *cobra.Command, args []string) {
		// try to load the environment
		name, _ := cmd.Flags().GetString("env")
		newenv, err := NewComfyEnvironmentFromExisting(name)
		if err != nil {
			if strings.HasPrefix(err.Error(), "environment not found") {
				fmt.Printf("Environment '%s' not found.  Create new environment with 'comfycli env create %s'\n", name, name)
			} else {
				slog.Error("error getting environment", "error", err)
			}
			os.Exit(1)
		}

		// run the little bugger
		newenv.Environment.BoundRunPythonScriptFromFile(filepath.Join(newenv.ComfyUIPath, "main.py"), args...)
	},
}

func InitRunComfy(envCmd *cobra.Command) {
	envCmd.AddCommand(runcomfyCmd)

	runcomfyCmd.PersistentFlags().String("env", "default", "Name of the environment to run ComfyUI")
}
