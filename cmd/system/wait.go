/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package system

import (
	"fmt"

	"github.com/spf13/cobra"
)

// waitCmd represents the wait command
var waitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Wait for a ComfyUI instance's queue to empty",
	Long:  `Wait for a ComfyUI instance's queue to empty`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("wait called")
	},
}

func InitWait(systemCmd *cobra.Command) {
	systemCmd.AddCommand(waitCmd)
}
