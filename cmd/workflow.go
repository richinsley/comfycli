/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package cmd

import (
	"log"

	"github.com/richinsley/comfycli/cmd/workflow"

	"github.com/spf13/cobra"
)

// workflowCmd represents the workflow command
var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Perform workflow operations with a ComfyUI instance",
	Long:  `Perform workflow operations with a ComfyUI instance`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				log.Fatalf("Error: %v", err)
			}
			return
		}
		// You can keep this or adjust as needed
		log.Println("workflow called with args: ", args)
	},
}

func init() {
	rootCmd.AddCommand(workflowCmd)

	// workflowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	CLIOptions.ApplyEnvironment()
	workflow.SetLocalOptions(&CLIOptions)

	workflow.InitParse(workflowCmd)
	workflow.InitQueue(workflowCmd)
	workflow.InitApi(workflowCmd)
	workflow.InitExtract(workflowCmd)
	workflow.InitInject(workflowCmd)
}
