/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package cmd

import (
	"log"

	utilities "github.com/richinsley/comfycli/cmd/util"
	"github.com/spf13/cobra"
)

// utilCmd represents the system command
var utilCmd = &cobra.Command{
	Use:   "util",
	Short: "Utility commands for Comfycli",
	Long:  `Utility commands for Comfycli`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				log.Fatalf("Error: %v", err)
			}
			return
		}
		// You can keep this or adjust as needed
		log.Println("util called with args: ", args)
	},
}

func init() {
	rootCmd.AddCommand(utilCmd)

	// hand over cli options to the system package
	CLIOptions.ApplyEnvironment()
	utilities.SetLocalOptions(&CLIOptions)

	// add system subcommands
	utilities.InitFileServe(utilCmd)
}
