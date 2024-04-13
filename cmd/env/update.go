/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package env

import (
	"fmt"
	"log/slog"
	"os"

	util "github.com/richinsley/comfycli/pkg"
	kinda "github.com/richinsley/kinda/pkg"
	"github.com/spf13/cobra"
)

// updateCmd
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an environment",
	Long:  `Update an environment`,
	Run: func(cmd *cobra.Command, args []string) {
		// get the list of environments
		envlist, err := GetComfyEnvironments()
		if err != nil {
			slog.Error("error getting environment list", "error", err)
			os.Exit(1)
		}

		if len(envlist) == 0 {
			fmt.Println("No environments found.  Create a new environment with 'comfycli env create <name>'")
			os.Exit(0)
		}

		if len(args) == 0 {
			if len(envlist) != 1 {
				slog.Error("error: no environment specified")
				os.Exit(1)
			}
			// default to the first environment
			args = append(args, envlist[0])
		}
		env := args[0]

		// check if the environment exists
		gotit := false
		for _, v := range envlist {
			if v == env {
				gotit = true
				break
			}
		}
		if !gotit {
			slog.Error("error: environment not found")
			os.Exit(1)
		}

		newenv, err := NewComfyEnvironmentFromExisting(env)
		if err != nil {
			slog.Error("error getting environment", "error", err)
			os.Exit(1)
		}

		if !CLIOptions.Yes {
			response, err := util.YesNo(fmt.Sprintf("Update environment %s: %s", env, newenv.Environment.EnvPath), true)
			if err != nil {
				slog.Error("error getting user response", "error", err)
				os.Exit(1)
			}
			if !response {
				os.Exit(0)
			}
		}

		// ask the user if they are sure

		// update the environment

		var outputstyle kinda.CreateEnvironmentOptions = kinda.ShowProgressBar
		if outputquiet {
			outputstyle = kinda.ShowNothing
		}
		if outputverbose {
			outputstyle = kinda.ShowProgressBarVerbose
		}
		if !outputquiet {
			fmt.Printf("Updating environment %s: %s\n", env, newenv.Environment.EnvPath)
		}

		err = newenv.UpdateEnvironment(outputstyle)
		if err != nil {
			slog.Error("error updating environment", "error", err)
			os.Exit(1)
		}
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		CheckForDefaultRecipe()
	},
}

func InitUpdate(envCmd *cobra.Command) {
	envCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&outputverbose, "verbose", "", false, "Verbose output")
	updateCmd.Flags().BoolVarP(&outputquiet, "quiet", "q", false, "Silent output")
}
