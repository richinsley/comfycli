/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// managerCmd represents the system command
var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "ComfyUI-Manager operations",
	Long:  `Perform operations with the ComfyUI-Manager system.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("system called")
	},
}

// TODO
func init() {
	// rootCmd.AddCommand(managerCmd)

	// // hand over cli options to the system package
	// CLIOptions.ApplyEnvironment()
	// manager.SetLocalOptions(&CLIOptions)

	// add system subcommands
	// manager.InitInfo(managerCmd)
}
