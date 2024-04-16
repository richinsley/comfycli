/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/richinsley/comfycli/cmd/env"
	"github.com/richinsley/comfycli/pkg"
	kinda "github.com/richinsley/kinda/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed version.txt
var version string

var CLIOptions = pkg.ComfyOptions{}
var ComfycliVersion kinda.Version

func PreprocessOptions(cmd *cobra.Command, args []string) {
	// parse host and port from the command line
	// if the host is in the form of host:port, split it
	host := viper.GetString("host")
	if host != "" {
		hostParts := strings.Split(host, ":")
		if len(hostParts) == 2 {
			CLIOptions.Host = hostParts[0]
			port, err := strconv.Atoi(hostParts[1])
			if err != nil {
				// handle the error if the conversion fails
				slog.Error("Failed to convert host port to integer:", "error", err)
				os.Exit(1)
			} else {
				CLIOptions.Port = port
			}
		} else {
			CLIOptions.Host = host
			CLIOptions.Port = 8188
		}
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "comfycli",
	Short:            "A command-line interface for interacting with ComfyUI",
	PreRun:           PreprocessOptions,
	PersistentPreRun: PreprocessOptions,
	Run: func(cmd *cobra.Command, args []string) {
		if CLIOptions.GetVersion {
			if CLIOptions.Json {
				versionmap := map[string]string{
					"version": ComfycliVersion.String(),
				}
				json, err := pkg.ToJson(versionmap, CLIOptions.PrettyJson)
				if err != nil {
					slog.Error("Failed to convert version to json:", "error", err)
					os.Exit(1)
				}
				fmt.Println(json)
				return
			}
			fmt.Println(ComfycliVersion.String())
			return
		}
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				slog.Error("Error:", "error", err)
				os.Exit(1)
			}
			return
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// setupComfycliHome tries to determine the comfycli home folder by looking for comfycli_config.yaml
// if createIfNeeded is true, it will create the folder structure in the path it does not exist.
func setupComfycliHome(path string, createIfNeeded bool) (string, error) {
	configpath := filepath.Join(path, "comfycli_config.yaml")
	if _, err := os.Stat(configpath); err == nil {
		// good to go I guess
		return configpath, nil
	}

	if createIfNeeded {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return "", err
		}
		// create a default config file
		viper.SafeWriteConfigAs(configpath)
		return configpath, nil
	}

	return "", fmt.Errorf("comfycli home folder not found at %s", path)
}

// checkComfycliHome checks the comfycli home folder structure and creates it if it needed
func checkComfycliHome(path string) error {
	/*
		HOME/
		├─ comfycli_config.yaml
		├─ models/
		├─ environments/
		│  ├─ recipes/
		│  │  ├─ default.json
	*/
	// check the models folder
	modelsFolder := filepath.Join(path, "models")
	if _, err := os.Stat(modelsFolder); os.IsNotExist(err) {
		fmt.Printf("Creating models folder: %s\n", modelsFolder)
		err := os.MkdirAll(modelsFolder, os.ModePerm)
		if err != nil {
			fmt.Printf("Failed to create models folder: %s\n", err)
			return err
		}
	}

	// create the recipes folder
	recipesFolder := filepath.Join(path, "environments", "recipes")
	if _, err := os.Stat(recipesFolder); os.IsNotExist(err) {
		fmt.Printf("Creating recipes folder: %s\n\n", recipesFolder)
		err := os.MkdirAll(recipesFolder, os.ModePerm)
		if err != nil {
			fmt.Printf("Failed to create recipes folder: %s\n", err)
			return err
		}
	}

	// create the recipes repos folder
	recipesReposFolder := filepath.Join(recipesFolder, "repos")
	if _, err := os.Stat(recipesReposFolder); os.IsNotExist(err) {
		fmt.Printf("Creating recipes repos folder: %s\n\n", recipesReposFolder)
		err := os.MkdirAll(recipesReposFolder, os.ModePerm)
		if err != nil {
			fmt.Printf("Failed to create recipes repos folder: %s\n", err)
			return err
		}
	}

	// populate the recipes folder with the default recipes
	recipes, err := env.GetEmeddedRecipeNames()
	if err == nil {
		for _, recipe := range recipes {
			// skip if the file already exists
			recipePath := filepath.Join(recipesFolder, recipe)
			if _, err := os.Stat(recipePath); os.IsNotExist(err) {
				embeddedRecipe, err := env.GetEmbeddedRecipe(recipe)
				if err != nil {
					fmt.Printf("Failed to get embedded recipe: %s\n", err)
					return err
				}
				err = embeddedRecipe.WriteRecipe(recipePath, false)
				if err != nil {
					fmt.Printf("Failed to write embedded recipe: %s\n", err)
					return err
				}
			}
		}
	}

	return nil
}

// GetComfycliHome trys to determine the comfycli home folder
// and creates it if it does not exist.
// The order of where it looks for the file comfycli_config.yaml is:
// 1. The directory specified by the environment variable COMFYCLI_HOME
// 2. The current working directory
// 3. $HOME/comfycli
func getComfycliHome() (string, error) {
	var configFilePath string
	var err error

	// If there is a COMFYCLI_HOME environment variable, try to use it
	// Because it's an explicit definition, we will create the home folder if it does not exist
	homedir := viper.GetString("HOME")
	if homedir != "" {
		configFilePath, err = setupComfycliHome(homedir, true)
		if err == nil && configFilePath != "" {
			// set up the home folder
			if err := checkComfycliHome(homedir); err != nil {
				return "", err
			}
			return configFilePath, nil
		}
	}

	// check the current directory
	currentdir, _ := os.Getwd()
	configFilePath, err = setupComfycliHome(currentdir, false)
	if err == nil && configFilePath != "" {
		// set up the home folder
		if err := checkComfycliHome(configFilePath); err != nil {
			return "", err
		}
		return configFilePath, nil
	}

	// last but not least, check the user's home directory
	userhomedir, _ := os.UserHomeDir()
	userhomedir = filepath.Join(userhomedir, ".comfycli")
	configFilePath, err = setupComfycliHome(userhomedir, true)
	if err == nil && configFilePath != "" {
		// set up the home folder
		if err := checkComfycliHome(userhomedir); err != nil {
			return "", err
		}
		return configFilePath, nil
	}

	// // Check if the work folder exists, create it if it does not.
	// if _, err := os.Stat(workFolder); os.IsNotExist(err) {
	// 	fmt.Printf("Creating comfycli home folder: %s\n", workFolder)
	// 	err := os.MkdirAll(workFolder, os.ModePerm)
	// 	if err != nil {
	// 		fmt.Printf("Failed to create comfycli home folder: %s\n", err)
	// 		os.Exit(1)
	// 	}
	// }

	return "", errors.New("comfycli home folder not found")
}

func getLongDescription() string {
	ComfycliVersion, _ = kinda.ParseVersion(version)
	return fmt.Sprintf(`A feature-rich command-line application designed to streamline 
the interaction with and scripting for ComfyUI for a shell.

Version: %s
Home Path: %s`, ComfycliVersion.String(), CLIOptions.HomePath)
}

func init() {
	viper.SetEnvPrefix("COMFYCLI") // Set the prefix for environment variables.
	viper.AutomaticEnv()           // Read in environment variables that match.

	// Determine the default work folder and check for an override from environment variables.
	defaultWorkFolder := filepath.Join(os.Getenv("HOME"), ".comfycli")
	workFolder := viper.GetString("HOME") // Looking for COMFYCLI_HOME
	if workFolder == "" {
		workFolder = defaultWorkFolder
	}

	// Check if the work folder exists, create it if it does not.
	if _, err := os.Stat(workFolder); os.IsNotExist(err) {
		fmt.Printf("Creating comfycli home folder: %s\n", workFolder)
		err := os.MkdirAll(workFolder, os.ModePerm)
		if err != nil {
			fmt.Printf("Failed to create comfycli home folder: %s\n", err)
			os.Exit(1)
		}
	}

	// Configure Viper to look for the configuration file in the work folder.
	// viper.AddConfigPath(workFolder)
	viper.SetConfigName("comfycli_config") // Name of the config file (without extension)
	viper.SetConfigType("yaml")            // REQUIRED if the config file does not have the extension in the name

	// Try to read the config or create it if it doesn't exist.
	configpath, err := getComfycliHome()
	if err != nil {
		slog.Error("Failed to get comfycli home folder", "error", err)
		os.Exit(1)
	}
	viper.SetConfigFile(configpath)
	if err := viper.ReadInConfig(); err != nil {
		slog.Error("Failed to create comfycli config file", "error", err)
		os.Exit(1)
	}

	// remove the config tile portion of the path
	CLIOptions.HomePath = filepath.Dir(configpath)
	CLIOptions.RecipesPath = filepath.Join(CLIOptions.HomePath, "environments", "recipes")
	CLIOptions.RecipesRepos = filepath.Join(CLIOptions.RecipesPath, "repos")
	CLIOptions.PrettyJson = true

	// add cobra subcommands
	rootCmd.PersistentFlags().StringVarP(&CLIOptions.Host, "host", "", "127.0.0.1:8188", "Host address")
	rootCmd.PersistentFlags().StringVarP(&CLIOptions.API, "api", "", "API", "Simple API title")
	rootCmd.PersistentFlags().StringVarP(&CLIOptions.APIValues, "apivalues", "", "", "Path to API values JSON or '-' for stdin")
	rootCmd.PersistentFlags().BoolVarP(&CLIOptions.Json, "json", "j", false, "Report all output as json")
	rootCmd.PersistentFlags().BoolVarP(&CLIOptions.DataToStdout, "stdout", "s", false, "Write node output data to stdout")
	rootCmd.PersistentFlags().BoolVarP(&CLIOptions.Yes, "yes", "y", false, "Automatically answer yes on prompted questions")
	rootCmd.PersistentFlags().BoolVarP(&CLIOptions.GetVersion, "version", "v", false, "Print the version of comfycli")

	// Set the Long field of rootCmd after CLIOptions.HomePath is populated
	rootCmd.Long = getLongDescription()

	// use viper to bind flags to config
	// this allows for automatically binding environment variables to registered parameters:
	// export COMFYCLI_HOST=192.168.0.51:8188
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
}
