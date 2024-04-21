/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package system

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/richinsley/comfy2go/client"
	"github.com/richinsley/comfycli/pkg"
	"github.com/spf13/cobra"
)

var interval int = 4

// topCmd represents the info command
var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Provides a dynamic real-time view of system information from a ComfyUI instance",
	Long:  `Provides a dynamic real-time view of system information from a ComfyUI instance`,
	Run: func(cmd *cobra.Command, args []string) {
		// create a client
		c := client.NewComfyClient(CLIOptions.Host[0], CLIOptions.Port[0], nil)

		// continuously update the top information at the specified interval
		for {
			// clear the terminal screen
			// windows
			if runtime.GOOS == "windows" {
				fmt.Print("\033[H\033[2J")
			} else {
				// unix
				fmt.Print("\033c")
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

			// sleep for the specified interval
			time.Sleep(time.Duration(interval) * time.Second)
		}
	},
}

func InitTop(systemCmd *cobra.Command) {
	topCmd.Flags().IntVarP(&interval, "interval", "i", 4, "Interval in seconds to update the top information")
	systemCmd.AddCommand(topCmd)
}
