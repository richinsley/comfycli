package env

import (
	"embed"
	"encoding/json"
	"os"
	"path"
	"runtime"

	"github.com/richinsley/comfycli/pkg"
)

const (
	// the current recipe format
	CurrentRecipeFormat  = "1.0"
	CurrentMinimumPython = "3.10"
)

/*
// base recipe
{
    "name": "default",
    "description": "Default macos environment for ComfyUI",
    "version": "1.0",
    "recipe_format": "1.0",
    "python": "3.10",
    "channel": "conda-forge",
    "pip_packages": [
        {
          "extra_index_url": "https://download.pytorch.org/whl/nightly/cpu",
          "packages": [
            {
              "name": "torch"
            },
            {
              "name": "torchvision"
            },
            {
              "name": "torchaudio"
            }
          ]
        }
    ],
    "models": [],
    "custom_nodes": [
      {
        "name": "Manager",
        "git_url": "https://github.com/ltdrdata/ComfyUI-Manager.git",
        "branch": "main"
      }
    ]
  }

  // inherited recipe
  {
    "name": "SD15",
    "description": "Stable Diffusion 1.5",
    "version": "1.0",
    "recipe_format": "1.0",
    "inherits": ["base"],
    "models": [
      {
        "name": "v1-5-pruned-emaonly.ckpt",
        "type": "checkpoints",
        "base": "SD1.5",
        "save_path": "default",
        "description": "Stable Diffusion 1.5 base model",
        "reference": "https://huggingface.co/runwayml/stable-diffusion-v1-5",
        "filename": "v1-5-pruned-emaonly.ckpt",
        "url": "https://huggingface.co/runwayml/stable-diffusion-v1-5/resolve/main/v1-5-pruned-emaonly.ckpt"
      }
    ]
  }

*/

type PipPackage struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type PipPackages struct {
	ExtraIndexURL string       `json:"extra_index_url,omitempty"`
	IndexURL      string       `json:"index_url,omitempty"`
	Packages      []PipPackage `json:"packages"`
}

type CustomNode struct {
	Name   string `json:"name"`
	GitURL string `json:"git_url"`
	Branch string `json:"branch,omitempty"`
}

// based off of ComfyUI-Manager format
//
//	{
//		"name": "v1-5-pruned-emaonly.ckpt",
//		"type": "checkpoints",
//		"base": "SD1.5",
//		"save_path": "default",
//		"description": "Stable Diffusion 1.5 base model",
//		"reference": "https://huggingface.co/runwayml/stable-diffusion-v1-5",
//		"filename": "v1-5-pruned-emaonly.ckpt",
//		"url": "https://huggingface.co/runwayml/stable-diffusion-v1-5/resolve/main/v1-5-pruned-emaonly.ckpt"
//	}
type Models struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Base        string `json:"base"`
	SavePath    string `json:"save_path"`
	Description string `json:"description,omitempty"`
	Reference   string `json:"reference,omitempty"`
	Filename    string `json:"filename"`
	URL         string `json:"url"`
}

type EnvRecipe struct {
	// default name for environment
	Name string `json:"name"`
	// description of the environment
	Description string `json:"description,omitempty"`
	// version of recipe json format
	RecipeFormat string `json:"recipe_format"`
	// version for the format of the recipe
	Version string `json:"version"`
	// python version to use
	PythonVersion string `json:"python,omitempty"`
	// list of recipes to inherit from
	Inherits []string `json:"inherits,omitempty"`
	// conda channel to use (optional)
	Channel *string `json:"channel,omitempty"`
	// pip packages to pre-install (optional)
	PipPackages []PipPackages `json:"pip_packages,omitempty"`
	// custom nodes to install (optional)
	CustomNodes []CustomNode `json:"custom_nodes,omitempty"`
	// models to install (optional)
	Models []Models `json:"models,omitempty"`
}

//go:embed recipes/darwin/* recipes/windows/* recipes/linux/* recipes/all/*
var EmbeddedRecipes embed.FS

func GetEmeddedRecipeNames() ([]string, error) {
	// get the recipes for the current OS
	rpath := path.Join("recipes", runtime.GOOS)
	recipes, err := EmbeddedRecipes.ReadDir(rpath)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(recipes))
	for i, recipe := range recipes {
		_, filename := path.Split(recipe.Name())
		name := filename
		names[i] = name
	}

	// get the all recipes
	rpath = path.Join("recipes", "all")
	recipes, err = EmbeddedRecipes.ReadDir(rpath)
	if err != nil {
		return nil, err
	}
	for _, recipe := range recipes {
		_, filename := path.Split(recipe.Name())
		name := filename
		names = append(names, name)
	}
	return names, nil
}

func GetEmbeddedRecipe(name string) (*EnvRecipe, error) {
	rpath := path.Join("recipes", runtime.GOOS, name)
	recipe, err := EmbeddedRecipes.ReadFile(rpath)
	if err != nil {
		// try all
		rpath = path.Join("recipes", "all", name)
		recipe, err = EmbeddedRecipes.ReadFile(rpath)
		if err != nil {
			return nil, err
		}
	}
	return ParseRecipe(recipe)
}

func (r *EnvRecipe) WriteRecipe(path string, overwrite bool) error {
	if _, err := os.Stat(path); err == nil && !overwrite {
		return os.ErrExist
	}
	recipe, err := pkg.ToJson(r, true)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(recipe), 0644)
}

func ParseRecipe(recipe []byte) (*EnvRecipe, error) {
	var envRecipe EnvRecipe
	err := json.Unmarshal(recipe, &envRecipe)
	if err != nil {
		return nil, err
	}
	return &envRecipe, nil
}
