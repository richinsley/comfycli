/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package cmd

import (
	"log"

	"github.com/richinsley/comfycli/cmd/system"
	"github.com/spf13/cobra"
)

// systemCmd represents the system command
var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "System commands for a ComfyUI instance",
	Long:  `System commands for a ComfyUI instance`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				log.Fatalf("Error: %v", err)
			}
			return
		}
		// You can keep this or adjust as needed
		log.Println("system called with args: ", args)
	},
}

func init() {
	rootCmd.AddCommand(systemCmd)

	// hand over cli options to the system package
	CLIOptions.ApplyEnvironment()
	system.SetLocalOptions(&CLIOptions)

	// add system subcommands
	system.InitInfo(systemCmd)
	system.InitCanrun(systemCmd)
	system.InitWait(systemCmd)
	system.InitNodes(systemCmd)
}
