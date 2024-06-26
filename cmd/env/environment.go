package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/go-git/go-git/v5"
	util "github.com/richinsley/comfycli/pkg"
	kinda "github.com/richinsley/kinda/pkg"
)

// create an environemnt from a recipe
type ComfyEnvironment struct {
	// name of the environment
	Name string `json:"Name"`
	// description of the environment
	Description string `json:"Description,omitempty"`
	// path to the recipe file
	RecipePath string `json:"RecipePath"`
	// python version used
	PythonVersion string `json:"PythonVersion"`
	// conda channel used
	Channel string `json:"Channel,omitempty"`
	// Path to the ComfyUI repository
	ComfyUIPath string `json:"ComfyUIPath"`
	// The environment
	Environment *kinda.Environment `json:"-"`
	// paramsets
	ParamSets map[string][]string `json:"paramsets,omitempty"`
	// Using shared models
	SharedModels bool `json:"SharedModels,omitempty"`
}

func GetComfyEnvironments() ([]string, error) {
	// create the environment path
	envpath := path.Join(CLIOptions.HomePath, "environments")
	environmentspath := path.Join(envpath, "envs")
	// ensure the environments folder exists, create it if it does not
	err := os.MkdirAll(environmentspath, 0755)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(path.Join(envpath, "envs"))
	if err != nil {
		return nil, err
	}
	retv := []string{}
	for _, v := range entries {
		retv = append(retv, v.Name())
	}
	return retv, nil
}

func NewComfyEnvironmentFromExisting(name string) (*ComfyEnvironment, error) {
	// create the environment
	envpath := path.Join(CLIOptions.HomePath, "environments")

	// ensure the environment exists
	envfile := path.Join(envpath, "envs", name, "kinda_env.json")
	if _, err := os.Stat(envfile); os.IsNotExist(err) {
		return nil, fmt.Errorf("environment not found: %s", name)
	}
	env, err := kinda.CreateEnvironment(name, envpath, "", "", kinda.ShowProgressBar)
	if err != nil {
		fmt.Printf("Error creating environment: %v\n", err)
		return nil, err
	}

	// load the env descriptor
	jdata, err := os.ReadFile(envfile)
	if err != nil {
		return nil, err
	}
	retv := &ComfyEnvironment{}
	err = json.Unmarshal(jdata, retv)
	if err != nil {
		return nil, err
	}
	retv.Environment = env
	return retv, nil
}

// create a new environment
func NewComfyEnvironmentFromRecipe(name string, recipe *EnvRecipe, recipePath string, feedback kinda.CreateEnvironmentOptions) (*ComfyEnvironment, error) {

	// create the environment
	envpath := path.Join(CLIOptions.HomePath, "environments")
	env, err := kinda.CreateEnvironment(name, envpath, recipe.PythonVersion, "conda-forge", feedback)
	if err != nil {
		fmt.Printf("Error creating environment: %v\n", err)
		return nil, err
	}

	// if the environment has a kinda_env.json already, then this env already exists
	envfile := path.Join(env.EnvPath, "kinda_env.json")
	if _, err := os.Stat(envfile); err == nil {
		return nil, fmt.Errorf("environment already exists")
	}

	// clone the comfyui repository into the environment
	if feedback == kinda.ShowVerbose || feedback == kinda.ShowProgressBarVerbose {
		fmt.Printf("Cloning ComfyUI repository\n")
	}

	cloneoptions := &git.CloneOptions{
		URL: "https://github.com/comfyanonymous/ComfyUI.git",
	}

	if feedback == kinda.ShowVerbose || feedback == kinda.ShowProgressBarVerbose {
		cloneoptions.Progress = os.Stdout
	}

	comfyFolder := filepath.Join(env.EnvPath, "comfyui")
	repo, err := git.PlainClone(comfyFolder, false, cloneoptions)

	if feedback != kinda.ShowNothing {
		fmt.Printf("Cloning ComfyUI repository\n")
	}

	if err != nil && err.Error() != "repository already exists" {
		fmt.Printf("Error cloning: %v\n", err)
		return nil, err
	}
	if repo == nil {
		return nil, fmt.Errorf("error cloning repository")
	}

	// pre-install any required packages
	if recipe.PipPackages != nil {
		for _, t := range recipe.PipPackages {
			if feedback != kinda.ShowNothing {
				fmt.Printf("Installing pip packages: %v\n", t)
			}
			if t.Packages != nil {
				for _, v := range t.Packages {
					err = env.PipInstallPackage(v.Name, t.IndexURL, t.ExtraIndexURL, false, feedback)
					if err != nil {
						fmt.Printf("Error installing requirements pip pre-requirements: %v\n", err)
						return nil, err
					}
				}
			}
		}
	}

	// install ComfyUI python requirements
	comfyReqPath := path.Join(comfyFolder, "requirements.txt")
	if _, err := os.Stat(comfyReqPath); err == nil {
		if feedback != kinda.ShowNothing {
			fmt.Println("Installing ComfyUI pip required packages:")
		}
		err = env.PipInstallRequirmements(comfyReqPath, feedback)
		if err != nil {
			fmt.Printf("Error installing requirements: %v\n", err)
			return nil, err
		}
	}

	// install custom nodes if specified
	if recipe.CustomNodes != nil {
		for _, v := range recipe.CustomNodes {
			if feedback != kinda.ShowNothing {
				fmt.Printf("Installing Custom Node: %v\n", v.Name)
			}
			repo, repoPath, err := kinda.NewGitRepo(v.GitURL, filepath.Join(comfyFolder, "custom_nodes"), v.Branch)
			if err != nil {
				fmt.Printf("Error cloning custom node: %v\n", err)
				return nil, err
			}
			fmt.Println(repo)

			// install the pip requirements if requirements.txt exists
			customReqPath := path.Join(repoPath, "requirements.txt")
			if _, err := os.Stat(customReqPath); err == nil {
				err = env.PipInstallRequirmements(customReqPath, feedback)
				if err != nil {
					fmt.Printf("Error installing requirements: %v\n", err)
					return nil, err
				}
			}
		}
	}

	/*
		When Model.SavePath is "default", the model will be saved to the default path for the model type:
		checkpoints
		configs
		embeddings
		loras
		upscale_models
		clip
		controlnet
		gligen
		style_models
		vae
		clip_vision
		diffusers
		hypernetworks
		unet
		vae_approx
	*/

	// install models if specified
	useshared := !CLIOptions.NoSharedModels
	if recipe.Models != nil {
		models_path := path.Join(comfyFolder, "models")
		shared_models_path := path.Join(CLIOptions.HomePath, "models")
		for _, m := range recipe.Models {
			var savepath string
			if m.SavePath == "default" {
				savepath = m.Type
			} else {
				savepath = m.SavePath
			}

			if useshared {
				savepath = path.Join(shared_models_path, savepath)
			} else {
				savepath = path.Join(models_path, savepath)
			}

			// ensure the save folder exists
			err = os.MkdirAll(savepath, 0755)
			if err != nil {
				fmt.Printf("Error model creating save path: %v\n", err)
				return nil, err
			}

			// check if the model already exists
			if _, err := os.Stat(path.Join(savepath, m.Filename)); err == nil {
				if feedback == kinda.ShowVerbose || feedback == kinda.ShowProgressBarVerbose {
					fmt.Printf("Model %s already exists at %s\n", m.Name, savepath)
				}
			} else {
				if feedback == kinda.ShowVerbose || feedback == kinda.ShowProgressBarVerbose {
					fmt.Printf("Downloading model %s to %s\n", m.Name, savepath)
				}

				savepath = path.Join(savepath, m.Filename)
				err = util.DownloadFile(m.URL, savepath, 5, feedback)
				if err != nil {
					fmt.Printf("Failed to download model %s: %v\n", m.Name, err)
					continue
				}
			}

			// if we downloaded a model, and we are using shared models, create a symlink from the shared models folder to the comfyui models folder
			if useshared {
				starget := m.SavePath
				if m.SavePath == "default" {
					starget = m.Type
				}
				sharedpath := path.Join(shared_models_path, starget, m.Filename)
				var savepath string
				if m.SavePath == "default" {
					savepath = m.Type
				} else {
					savepath = m.SavePath
				}
				truepath := path.Join(models_path, savepath, m.Filename)
				// ensure the path that we will put the symlink in exists
				err = os.MkdirAll(path.Dir(truepath), 0755)
				if err != nil {
					fmt.Printf("Error creating symlink path: %v\n", err)
					return nil, err
				}

				// if we are on windows, we need to use a hard link instead of a symlink
				if runtime.GOOS == "windows" {
					err = os.Link(sharedpath, truepath)
					if err != nil {
						fmt.Printf("Error creating hard link: %v\n", err)
						return nil, err
					}
				} else {
					// create the symlink
					err = os.Symlink(sharedpath, truepath)
					if err != nil {
						fmt.Printf("Error creating symlink: %v\n", err)
						return nil, err
					}
				}
			}
		}
	}

	if recipePath == "" {
		// write the recipe to the file to the same folder as the environment
		recipePath = path.Join(env.EnvPath, name+".json")
		err = recipe.WriteRecipe(recipePath, true)
		if err != nil {
			fmt.Printf("Error writing recipe: %v\n", err)
			return nil, err
		}
	}

	retv := &ComfyEnvironment{
		Name:          name,
		Description:   recipe.Description,
		RecipePath:    recipePath,
		PythonVersion: recipe.PythonVersion,
		Channel:       "",
		Environment:   env,
		ComfyUIPath:   comfyFolder,
		ParamSets:     recipe.ParamSets,
		SharedModels:  useshared,
	}

	if recipe.Channel != nil {
		retv.Channel = *recipe.Channel
	}

	// write the environment to disk as a json file
	jdata, err := util.ToJson(retv, true)
	if err != nil {
		return nil, err
	}

	// write the environment to disk
	err = os.WriteFile(envfile, []byte(jdata), 0644)
	if err != nil {
		return nil, err
	}

	return retv, nil
}

func (c *ComfyEnvironment) DeleteEnvironment() error {
	// delete the environment
	err := os.RemoveAll(c.Environment.EnvPath)
	if err != nil {
		return err
	}

	return nil
}

func (c *ComfyEnvironment) UpdateEnvironment(feedback kinda.CreateEnvironmentOptions) error {
	// update the environment

	// update the comfyui repository
	if feedback == kinda.ShowVerbose || feedback == kinda.ShowProgressBarVerbose || feedback == kinda.ShowProgressBar {
		fmt.Printf("Updating ComfyUI repository\n")
	}

	comfyFolder := filepath.Join(c.Environment.EnvPath, "comfyui")
	repo, err := git.PlainOpen(comfyFolder)
	if err != nil {
		return err
	}

	// use git to do a git pull
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err.Error() != "already up-to-date" {
		return err
	}

	if err == nil {
		// pip install any new requirements in the comfyui requirements.txt
		comfyReqPath := path.Join(comfyFolder, "requirements.txt")
		if _, err := os.Stat(comfyReqPath); err == nil {
			if feedback != kinda.ShowNothing {
				fmt.Println("Installing ComfyUI pip required packages:")
			}
			err = c.Environment.PipInstallRequirmements(comfyReqPath, feedback)
			if err != nil {
				return fmt.Errorf("error installing requirements: %v", err)
			}
		}
	} else if feedback == kinda.ShowVerbose || feedback == kinda.ShowProgressBarVerbose || feedback == kinda.ShowProgressBar {
		fmt.Printf("ComfyUI repository already up-to-date\n")
	}

	// get the list of all folders in the comfui custom_nodes folder
	customNodesPath := path.Join(comfyFolder, "custom_nodes")
	entries, err := os.ReadDir(customNodesPath)
	if err != nil {
		return err
	}

	// for each folder in the custom_nodes folder, check if it is a git repo, and if it is, do a git pull
	for _, v := range entries {
		if v.IsDir() {
			// dir entry to path string
			repopath := path.Join(customNodesPath, v.Name())
			repo, err := git.PlainOpen(repopath)
			if err != nil {
				// not a git repo
				continue
			}

			if feedback == kinda.ShowVerbose || feedback == kinda.ShowProgressBarVerbose || feedback == kinda.ShowProgressBar {
				fmt.Printf("Updating custom node repository: %s\n", v.Name())
			}

			// use git to do a git pull
			w, err := repo.Worktree()
			if err != nil {
				return err
			}

			err = w.Pull(&git.PullOptions{RemoteName: "origin"})
			if err != nil && err.Error() != "already up-to-date" {
				return err
			}

			if err == nil {
				// pip install any new requirements in the comfyui requirements.txt
				customReqPath := path.Join(customNodesPath, v.Name(), "requirements.txt")
				if _, err := os.Stat(customReqPath); err == nil {
					err = c.Environment.PipInstallRequirmements(customReqPath, feedback)
					if err != nil {
						return fmt.Errorf("error installing requirements: %v", err)
					}
				}
			} else if feedback == kinda.ShowVerbose || feedback == kinda.ShowProgressBarVerbose || feedback == kinda.ShowProgressBar {
				fmt.Printf("Custom node repository %s already up-to-date\n", v.Name())
			}
		}
	}

	return nil
}
