/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package env

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/spf13/cobra"
)

// lsenvCmd
var lsenvCmd = &cobra.Command{
	Use:   "ls",
	Short: "List available environments",
	Long:  `List available environments`,
	Run: func(cmd *cobra.Command, args []string) {
		// try to load the environment
		envpath := path.Join(CLIOptions.HomePath, "environments")
		fmt.Println(envpath)
		envlist, err := GetComfyEnvironments()
		if err != nil {
			slog.Error("error getting environment list", "error", err)
			os.Exit(1)
		}

		for _, v := range envlist {
			fmt.Println(v)
		}
	},
}

func InitLSEnv(envCmd *cobra.Command) {
	envCmd.AddCommand(lsenvCmd)
}
