/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package system

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/richinsley/comfy2go/client"
	"github.com/richinsley/comfycli/pkg"
	"github.com/spf13/cobra"
)

func displayAvailableNodes(c *client.ComfyClient) {
	object_infos, err := c.GetObjectInfos()
	if err != nil {
		slog.Error("Error decoding Object Infos:", err)
		os.Exit(1)
	}

	if CLIOptions.Json {
		j, err := pkg.ToJson(object_infos, CLIOptions.PrettyJson)
		if err != nil {
			slog.Error("Error fomating system info to json:", err)
			os.Exit(1)
		}
		fmt.Println(j)
	} else {
		fmt.Println("Available Nodes:")
		for _, n := range object_infos.Objects {
			fmt.Printf("\tNode Name: \"%s\"\n", n.DisplayName)
			props := n.GetSettableProperties()
			fmt.Printf("\t\tProperties:\n")
			for _, p := range props {
				fmt.Printf("\t\t\t\"%s\"\tType: [%s]\n", p.Name(), p.TypeString())
				if p.TypeString() == "COMBO" {
					c, _ := p.ToComboProperty()
					for _, combo_item := range c.Values {
						fmt.Printf("\t\t\t\t\"%s\"\n", combo_item)
					}
				}
			}
		}
	}
}

// nodesCmd represents the nodes command
var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List available nodes in a ComfyUI instance",
	Long:  `List available nodes in a ComfyUI instance`,
	Run: func(cmd *cobra.Command, args []string) {
		// create a client
		c := client.NewComfyClient(CLIOptions.Host[0], CLIOptions.Port[0], nil)

		displayAvailableNodes(c)
	},
}

func InitNodes(systemCmd *cobra.Command) {
	systemCmd.AddCommand(nodesCmd)
}
