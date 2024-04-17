/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package workflow

import (
	"fmt"
	"os"

	util "github.com/richinsley/comfycli/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api [workflow file]",
	Short: "Output the API for the workflow in json format",
	Long:  `Output the API for the workflow in json format`,
	// validate that a png file is provided
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("requires a workflow file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		workflowPath := args[0]
		params := args[1:] // All other args are considered parameters
		parameters := util.ParseParameters(params)

		_, graph, simple_api, missing, err := util.GetFullWorkflow(CLIOptions, workflowPath, nil)
		if missing != nil {
			slog.Error("failed to get workflow: missing nodes", missing)
			os.Exit(1)
		}

		if err != nil {
			slog.Error("failed to get workflow", err)
			os.Exit(1)
		}

		if CLIOptions.APIValuesOnly {
			// create a slice of the API parameter values and serialize to json
			_, err = util.ApplyParameters(nil, CLIOptions, graph, simple_api, parameters)
			if err != nil {
				slog.Error("failed to apply parameter", err)
				os.Exit(1)
			}

			values := make(map[string]interface{})
			for k, v := range simple_api.Properties {
				values[k] = v.GetValue()
			}

			j, _ := util.ToJson(values, CLIOptions.PrettyJson)
			fmt.Println(j)
		} else {
			// serialize the entire API to json
			j, _ := util.ToJson(simple_api.Properties, CLIOptions.PrettyJson)
			fmt.Println(j)
		}
	},
}

func InitApi(workflowCmd *cobra.Command) {
	workflowCmd.AddCommand(apiCmd)

	apiCmd.Flags().BoolVarP(&CLIOptions.APIValuesOnly, "values", "", false, "Output as values only")
}
