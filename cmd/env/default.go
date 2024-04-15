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

// defaultCMD
var defaultCMD = &cobra.Command{
	Use:   "setdefault",
	Short: "Set the default base recipe for the target system",
	Long: `Set the default base recipe for the target system.
Different archectures and GPUs may require different base recipes.
The base recipe is the root recipe that all other recipes will inherit from.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			slog.Error("error: no environment specified")
			os.Exit(1)
		}
		recipe := args[0]

		// get list of recipes from the home folder
		if CLIOptions.RecipesPath == "" {
			fmt.Println("recipes path not set")
			return
		}

		// find the recipes with "default_" prefix
		defaults, err := GetAllDefaultRecipes()
		if err != nil {
			fmt.Println("error getting default recipes")
			return
		}

		// if no default recipes are found, exit
		if len(defaults) == 0 {
			fmt.Println("no default recipes found")
			return
		}

		// check if the recipe exists
		gotit := false
		for k := range defaults {
			if k == recipe {
				gotit = true
				break
			}
		}
		if !gotit {
			fmt.Println("error: recipe not found")
			return
		}

		// set the default recipe
		newdefaultpath := defaults[recipe]
		// we need to deserialize the recipe, change the name to 'default' and write it to 'default.json'
		r, err := RecipeFromPath(newdefaultpath)
		if err != nil {
			fmt.Printf("error deserializing recipe %s: %s\n", recipe, err)
			os.Exit(1)
		}
		r.Name = "default"
		err = r.WriteRecipe(path.Join(CLIOptions.RecipesPath, "default.json"), true)
		if err != nil {
			fmt.Printf("error writing default recipe %s: %s\n", recipe, err)
			os.Exit(1)
		}
		fmt.Printf("setting default recipe to %s\n", recipe)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		CheckForDefaultRecipe()
	},
}

func InitDefault(envCmd *cobra.Command) {
	envCmd.AddCommand(defaultCMD)
}
