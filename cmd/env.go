/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package cmd

import (
	"log"

	"github.com/richinsley/comfycli/cmd/env"
	"github.com/spf13/cobra"
)

// envCmd represents the system command
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Create and manage python virtual environments for ComfyUI",
	Long:  `Create and manage python virtual environments for ComfyUI`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				log.Fatalf("Error: %v", err)
			}
			return
		}
		// You can keep this or adjust as needed
		log.Println("env called with args: ", args)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		env.CheckForDefaultRecipe()
	},
}

func init() {
	rootCmd.AddCommand(envCmd)

	// hand over cli options to the system package
	CLIOptions.ApplyEnvironment()
	env.SetLocalOptions(&CLIOptions)

	// add env subcommands
	env.InitCreate(envCmd)
	env.InitRecipes(envCmd)
	env.InitRunComfy(envCmd)
	env.InitLSEnv(envCmd)
	env.InitRMEnv(envCmd)
	env.InitUpdate(envCmd)
}
