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

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Retrieve system information from a ComfyUI instance",
	Long:  `Retrieve system information from a ComfyUI instance`,
	Run: func(cmd *cobra.Command, args []string) {
		// create a client
		c := client.NewComfyClient(CLIOptions.Host, CLIOptions.Port, nil)

		// the client needs to be in an initialized state before usage
		if !c.IsInitialized() {
			err := c.Init()
			if err != nil {
				slog.Error("Error initializing client:", "error", err)
				os.Exit(1)
			}
		}

		s, err := c.GetSystemStats()
		if err != nil {
			slog.Error("Error initializing client:", "error", err)
			os.Exit(1)
		}

		if CLIOptions.Json {
			j, err := pkg.ToJson(s, CLIOptions.PrettyJson)
			if err != nil {
				slog.Error("Error fomating system info to json:", err)
				os.Exit(1)
			}
			fmt.Println(j)
		} else {
			// output system info as plain text
			fmt.Println("System Info:")
			fmt.Println("  OS:", s.System.OS)
			fmt.Println("  Python Version:", s.System.PythonVersion)
			fmt.Println("  Embedded Python:", s.System.EmbeddedPython)
			fmt.Println("Devices:")
			for _, d := range s.Devices {
				fmt.Println("  Name:", d.Name)
				fmt.Println("  Type:", d.Type)
				fmt.Println("  Index:", d.Index)
				fmt.Println("  VRAM Total:", d.VRAM_Total)
				fmt.Println("  VRAM Free:", d.VRAM_Free)
				fmt.Println("  Torch VRAM Total:", d.Torch_VRAM_Total)
				fmt.Println("  Torch VRAM Free:", d.Torch_VRAM_Free)
			}
		}
	},
}

func InitInfo(systemCmd *cobra.Command) {
	systemCmd.AddCommand(infoCmd)
}
