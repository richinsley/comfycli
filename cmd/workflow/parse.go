/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package workflow

import (
	"fmt"
	"os"

	util "github.com/richinsley/comfycli/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

// parseCmd represents the info command
var parseCmd = &cobra.Command{
	Use:   "parse [workflow file]",
	Short: "Parse a workflow file and output the workflow json.",
	Long: `Parse a workflow file and output the workflow json.
This allows you to modify the parameters of a workflow and output the json
to a file or the terminal.  Parameters follow the delimiter "--" and are in the
format "node:parameter"=value.  For example, to set the seed parameter of a KSampler node
to 1234, you would use "KSampler:seed"=1234

examples:
# parse the default workflow, set the KSampler seed parameter to 1234 and output the workflow json to a file
comfycli workflow parse defaultworkflow.json -- "KSampler:seed"=1234 > newworkflow.json
`,
	Run: func(cmd *cobra.Command, args []string) {
		workflowPath := args[0]
		params := args[1:] // All other args are considered parameters
		parameters := util.ParseParameters(params)

		workflow, _, missing, err := util.ClientWithWorkflow(0, CLIOptions, workflowPath, parameters, nil)
		if missing != nil {
			slog.Error("failed to get workflow: missing nodes", "missing", fmt.Sprintf("%v", missing))
			os.Exit(1)
		}

		if err != nil {
			slog.Error("error getting client and workflow", err)
			os.Exit(1)
		}

		j, _ := util.ToJson(workflow.Graph, CLIOptions.PrettyJson)
		if err != nil {
			slog.Error("failed to convert graph to json", err)
			os.Exit(1)
		}

		// queue the prompt and get the resulting image
		if CLIOptions.GraphOutPath != "" {
			d := []byte(j)
			err := util.SaveData(&d, CLIOptions.GraphOutPath)
			if err != nil {
				slog.Error("failed to save graph to file", err)
				os.Exit(1)
			}
		} else {
			fmt.Println(j)
		}
	},
}

func InitParse(workflowCmd *cobra.Command) {
	workflowCmd.AddCommand(parseCmd)

	parseCmd.PersistentFlags().StringVarP(&CLIOptions.GraphOutPath, "graphout", "g", "", "Path to write workflow graph JSON")
}
