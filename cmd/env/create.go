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

	util "github.com/richinsley/comfycli/pkg"
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
	comfycli env create -recipe default -python 3.10 -name myenv
	
	# combine multiple recipes into a new environment
	comfycli env create -recipe default,SD15,SDXL -name all_sd`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			slog.Error("invalid arguments", "args", args)
			// print the help
			cmd.Help()
			os.Exit(1)
		}
		recipe, _ := cmd.Flags().GetString("recipe")
		python, _ := cmd.Flags().GetString("python")
		name, _ := cmd.Flags().GetString("name")
		file, _ := cmd.Flags().GetString("file")
		var r *EnvRecipe = nil
		var err error

		// split the recipe string on commas
		// this allows for multiple recipes to be specified
		wanted_recipe_list := strings.Split(recipe, ",")

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
			// see if we have a 'default' recipe
			if !HasDefaultRecipe() {
				// we are almost certainly going to need a default system recipe
				defaults, err := GetAllDefaultRecipes()
				if err != nil {
					slog.Error("error loading default recipes", "error", err)
					os.Exit(1)
				}

				// check if a default recipe is in the wanted list
				default_candidate := ""
				default_candidate_path := ""
				for _, wanted := range wanted_recipe_list {
					for k, _ := range defaults {
						if wanted == k {
							// we have a default recipe in the wanted list
							// we can use it as the default recipe
							default_candidate = k
							default_candidate_path = path.Join(CLIOptions.RecipesPath, default_candidate+".json")
							break
						}
					}
				}

				if default_candidate != "" {
					// make default_candidate the default
					// we need to deserialize the recipe, change the name to 'default' and write it to 'default.json'
					r, err = RecipeFromPath(default_candidate_path)
					if err != nil {
						slog.Error("error deserializing recipe", "error", err)
						os.Exit(1)
					}
					r.Name = "default"
					err = r.WriteRecipe(path.Join(CLIOptions.RecipesPath, "default.json"), true)
					if err != nil {
						slog.Error("error writing default recipe", "error", err)
						os.Exit(1)
					}
				}
			}

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

		// present the recipe to the user, the target python version, the name of the environment, and the target path
		// ask the user if they want to proceed
		if !CLIOptions.Yes {
			fmt.Println("Creating ComfyUI environment with the following settings:")
			fmt.Printf("Recipe: %s\n", r.Name)
			fmt.Printf("Python: %s\n", r.PythonVersion)
			fmt.Printf("Name: %s\n", name)
			fmt.Printf("Path: %s\n", path.Join(CLIOptions.HomePath, "environments", "envs", name))
			proceed, err := util.YesNo("Proceed with environment creation?", true)
			if err != nil {
				slog.Error("invalid response", "error", err)
				os.Exit(1)
			}
			if !proceed {
				os.Exit(0)
			}
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
	PreRun: func(cmd *cobra.Command, args []string) {
		CheckForDefaultRecipe()
	},
}

func InitCreate(envCmd *cobra.Command) {
	envCmd.AddCommand(createCmd)

	createCmd.Flags().String("recipe", "default", "Recipe to use for environment creation")
	createCmd.Flags().String("python", "", "Override the recipe python version to use")
	createCmd.Flags().String("name", "default", "Name of the environment to create")
	createCmd.Flags().String("file", "", "Path to an external recipe file to use for environment creation")
	createCmd.Flags().BoolVarP(&outputverbose, "verbose", "", false, "Verbose output")
	createCmd.Flags().BoolVarP(&outputquiet, "quiet", "q", false, "Silent output")
	createCmd.Flags().BoolVarP(&CLIOptions.NoSharedModels, "noshared", "n", false, "Do not use shared models path")
}
