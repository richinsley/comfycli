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
	Use:   "queue [workflow file]",
	Short: "Queue a workflow for processing",
	Long: `
Queue a workflow for processing. The first argument is the path to the workflow file.
Set the parameters for the workflow by adding them as additional arguments after "--"
Node parameters are set by providing the node name followed by the parameter name and value.
When using a Simple API, parameters can be set by providing the parameter name and value.
Nodes that output data save the data to the current working directory.

examples:

# Set the seed parameter for a node with the title "KSampler"
comfycli workflow queue myworkflow.json -- KSampler:seed=1234

# Use a workflow that has a Simple API that has a parameter named "seed"
comfycli --api API workflow queue myworkflow_simple_api.json -- seed=1234

# Queue a workflow, don't save images to disk, but output them to the terminal using the Inline Image Protocol
comfycli workflow queue --inlineimages --nosavedata myworkflow.json -- KSampler:seed=1234
`,
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
	queueCmd.Flags().BoolVarP(&CLIOptions.InlineImages, "inlineimages", "i", false, "Output images to terminal with Inline Image Protocol")
	queueCmd.Flags().BoolVarP(&CLIOptions.NoSaveData, "nosavedata", "n", false, "Do not save data to disk")
	queueCmd.Flags().StringVarP(&CLIOptions.OutputNodes, "outputnodes", "o", "", "Specify which output nodes save data. Comma separated nodes. Default is all nodes")

	workflowCmd.AddCommand(queueCmd)
}
