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
	Use:   "parse",
	Short: "parse a workflow file and output the workflow json",
	Long:  `parse a workflow file and output the workflow json`,
	Run: func(cmd *cobra.Command, args []string) {
		workflowPath := args[0]
		params := args[1:] // All other args are considered parameters
		parameters := util.ParseParameters(params)

		_, graph, _, _, missing, err := util.ClientWithWorkflow(CLIOptions, workflowPath, parameters, nil)
		if missing != nil {
			slog.Error("failed to get workflow: missing nodes", "missing", fmt.Sprintf("%v", missing))
			os.Exit(1)
		}

		if err != nil {
			slog.Error("error getting client and workflow", err)
			os.Exit(1)
		}

		j, _ := util.ToJson(graph, CLIOptions.PrettyJson)
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
}
