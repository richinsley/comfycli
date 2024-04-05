/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package system

import (
	"fmt"
	"os"

	"github.com/richinsley/comfy2go/client"
	"github.com/richinsley/comfy2go/graphapi"
	util "github.com/richinsley/comfycli/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
)

func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

type missingComboValue struct {
	NodeTitle     string
	NodeType      string
	PropertyName  string
	PropertyValue string
}

func getMissingCombos(c *client.ComfyClient, graph *graphapi.Graph) []missingComboValue {
	object_infos, err := c.GetObjectInfos()
	if err != nil {
		slog.Error("Error decoding Object Infos:", "error", err)
		os.Exit(1)
	}
	missing := make([]missingComboValue, 0)
	for _, n := range graph.Nodes {
		for _, p := range n.Properties {
			obj, ok := object_infos.Objects[n.Type]
			if !ok && n.Type != "PrimitiveNode" && n.Type != "Note" && n.Type != "Reroute" {
				slog.Error("could not find node type:", "node", n.Type)
				os.Exit(1)
			}

			if p.TypeString() == "COMBO" {
				combo, _ := p.ToComboProperty()
				cvalue, ok := combo.GetValue().(string)
				if !ok {
					// combo is not a string value
					continue
				}
				inputrawprop := obj.InputPropertiesByID[combo.Name()]
				inputcomboprop, _ := (*inputrawprop).ToComboProperty()
				inputcombovalues := inputcomboprop.Values

				// check is cvalue is in inputcombovalues
				if !contains(inputcombovalues, cvalue) {
					mvalue := missingComboValue{
						NodeTitle:     n.DisplayName,
						NodeType:      n.Type,
						PropertyName:  p.Name(),
						PropertyValue: p.GetValue().(string),
					}
					missing = append(missing, mvalue)
				}
			}
		}
	}
	return missing
}

// canrunCmd represents the canrun command
var canrunCmd = &cobra.Command{
	Use:   "canrun [workflow file path]",
	Short: "Tests if an instance of ComfyUI can run a workflow with the given parameters",
	Long: `Tests if an instance of ComfyUI can run a workflow with the given parameters.
	Returns true if the instance can run the workflow, or a list of missing nodes and missing combo values if it cannot.
	examples:
	# test if the instance of ComfyUI at 192.168.0.41:9000 can run the workflow 'workflow.json'
	comfycli --host 192.168.0.41 --port 9000 system canrun /path/to/workflow.json`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			if err := cmd.Help(); err != nil {
				slog.Error("Error: %v", "error", err)
			}
			os.Exit(1)
		}

		workflowPath := args[0]
		params := args[1:] // All other args are considered parameters
		parameters := util.ParseParameters(params)

		// error outputs should go to stout instead of stderr
		c, graph, _, _, missing, err := util.ClientWithWorkflow(CLIOptions, workflowPath, parameters, nil)
		if err != nil && missing == nil {
			// is err a fs.PathError?
			if err, ok := err.(*os.PathError); ok {
				slog.Error("workflow not found", "error", err)
				os.Exit(1)
			}

			slog.Error("error getting client and workflow", "error", err)
			os.Exit(1)
		}

		if missing != nil {
			if !CLIOptions.Json {
				fmt.Println("failed to get workflow\nmissing nodes:\n--------------")
				for _, v := range *missing {
					fmt.Println(v)
				}
			} else {
				// output as json
				output := make(map[string]interface{})
				output["canrun"] = false
				output["missing_nodes"] = missing
				j, _ := util.ToJson(output, CLIOptions.PrettyJson)
				fmt.Println(j)
			}
			// we can't continue on to check for missing combo values
			os.Exit(0)
		}

		missingcombos := getMissingCombos(c, graph)
		if len(missingcombos) > 0 {
			if !CLIOptions.Json {
				fmt.Println("failed to get workflow\nmissing combo values:\n--------------")
				for _, v := range missingcombos {
					fmt.Println(v)
				}
			} else {
				// output as json
				output := make(map[string]interface{})
				output["canrun"] = false
				output["missing_combo_values"] = missingcombos
				j, _ := util.ToJson(output, CLIOptions.PrettyJson)
				fmt.Println(j)
			}
		} else {
			if !CLIOptions.Json {
				fmt.Printf("Host %s:%d can run workflow %s\n", CLIOptions.Host, CLIOptions.Port, workflowPath)
			} else {
				// output as json
				output := make(map[string]interface{})
				output["canrun"] = true
				j, _ := util.ToJson(output, CLIOptions.PrettyJson)
				fmt.Println(j)
			}
		}
	},
}

func InitCanrun(systemCmd *cobra.Command) {
	systemCmd.AddCommand(canrunCmd)
}
