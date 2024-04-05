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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("workflow called")
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
}
