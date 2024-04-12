/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package env

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// rmenvCmd
var rmenvCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove an environment",
	Long:  `Remove an environment`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			slog.Error("error: no environment specified")
			os.Exit(1)
		}
		env := args[0]

		// get the list of environments
		envlist, err := GetComfyEnvironments()
		if err != nil {
			slog.Error("error getting environment list", "error", err)
			os.Exit(1)
		}

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

		// remove the environment
		fmt.Printf("Removing environement %s: %s\n", env, newenv.Environment.EnvPath)

		// remove the entire newenv.Environment.EnvPath directory
		err = newenv.DeleteEnvironment()
		if err != nil {
			slog.Error("error removing environment", "error", err)
			os.Exit(1)
		}
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		CheckForDefaultRecipe()
	},
}

func InitRMEnv(envCmd *cobra.Command) {
	envCmd.AddCommand(rmenvCmd)
}
