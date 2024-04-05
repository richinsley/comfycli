/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package workflow

import (
	"github.com/richinsley/comfycli/pkg"
	"github.com/spf13/cobra"
)

// queueCmd represents the queue command
var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		workflowPath := args[0]
		params := args[1:] // All other args are considered parameters
		parameters := pkg.ParseParameters(params)

		// Process the queue. If there was a pipe loop, process it again
		for {
			hasPipeLoop := pkg.ProcessQueue(CLIOptions, workflowPath, parameters)
			if !hasPipeLoop {
				break
			}
		}
	},
}

func InitQueue(workflowCmd *cobra.Command) {
	workflowCmd.AddCommand(queueCmd)
}
