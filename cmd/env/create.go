/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package env

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	kinda "github.com/richinsley/kinda/pkg"
	"github.com/spf13/cobra"
)

var outputquiet bool
var outputverbose bool

// createCmd
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new virtual ComfyUI python environment",
	Long: `Create a new virtual ComfyUI python environment
	examples:
	# create a new environment using the default system recipe with the name default
	comfycli env create

	# create a new environment using the default system recipe with the name myenv
	comfycli env create -name myenv

	# create a new environment using the default recipe, force it to be python 3.10 and name it myenv
	comfycli env create -recipe default -python 3.10 -name myenv`,
	Run: func(cmd *cobra.Command, args []string) {
		recipe, _ := cmd.Flags().GetString("recipe")
		python, _ := cmd.Flags().GetString("python")
		name, _ := cmd.Flags().GetString("name")
		file, _ := cmd.Flags().GetString("file")
		var r *EnvRecipe = nil
		var err error

		// get list of all available recipes
		// recipes, err := GetRecipeList(CLIOptions.RecipesPath)
		// if err != nil {
		// 	slog.Error("error getting recipe list", "error", err)
		// 	return
		// }

		if file != "" {
			// check if the file exists
			if _, err := os.Stat(file); os.IsNotExist(err) {
				slog.Error("recipe file not found", "file", file)
				os.Exit(1)
			}

			r, err = RecipeFromPath(recipe)
			if err != nil {
				slog.Error("error deserializing recipe", "error", err)
				os.Exit(1)
			}
		} else {
			// split the recipe string on commas
			// this allows for multiple recipes to be specified
			wanted_recipe_list := strings.Split(recipe, ",")

			r, err = RecipeFromNames(wanted_recipe_list)
			if err != nil {
				slog.Error("error deserializing recipe", "error", err)
				os.Exit(1)
			}

			// override the python version
			if python != "" {
				r.PythonVersion = python
			}

			// if we have multiple wanted recipes, we've created a new one
			// check if we need to write it to a file.
			if len(wanted_recipe_list) > 1 {
				// check our recipe path for the recipe
				newrecipename := name
				if name == "default" {
					newrecipename = strings.Join(wanted_recipe_list, "-")
				}
				_, err := RecipePathFromName(newrecipename)
				if err != nil {
					recipe_path := path.Join(CLIOptions.RecipesPath, newrecipename+".json")
					// write the recipe to the file
					err = r.WriteRecipe(recipe_path, true)
					if err != nil {
						slog.Warn("error writing merged recipe", "error", err)
					}
				}
			}
		}

		if r == nil {
			slog.Error("recipe not found", "recipe", recipe)
			os.Exit(1)
		}

		// override the python version
		if python != "" {
			r.PythonVersion = python
		}

		// create the environment
		var outputstyle kinda.CreateEnvironmentOptions = kinda.ShowProgressBar
		if outputquiet {
			outputstyle = kinda.ShowNothing
		}
		if outputverbose {
			outputstyle = kinda.ShowProgressBarVerbose
		}
		env, err := NewComfyEnvironmentFromRecipe(name, r, file, outputstyle)
		if err != nil {
			slog.Error("error creating environment", "error", err)
			os.Exit(1)
		}
		if env != nil {
			fmt.Printf("Created environment: %s\n", env.Name)
		}
	},
}

func InitCreate(envCmd *cobra.Command) {
	envCmd.AddCommand(createCmd)

	createCmd.Flags().String("recipe", "default", "Recipe to use for environment creation")
	createCmd.Flags().String("python", "", "Override the recipe python version to use")
	createCmd.Flags().String("name", "default", "Name of the environment to create")
	createCmd.Flags().String("file", "", "Path to an external recipe file to use for environment creation")
	createCmd.Flags().BoolVarP(&outputverbose, "verbose", "v", false, "Verbose output")
	createCmd.Flags().BoolVarP(&outputquiet, "quiet", "q", false, "Silent output")
}
