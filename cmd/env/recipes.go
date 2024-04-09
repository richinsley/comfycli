/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package env

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	dagger "github.com/richinsley/comfycli/dagger"
	util "github.com/richinsley/comfycli/pkg"
	kinda "github.com/richinsley/kinda/pkg"
	"github.com/spf13/cobra"
)

func GetRecipeList(recipesPath string) ([]string, error) {
	// get list of recipe files from the recipes folder
	recipeFiles, err := util.ListFiles(recipesPath)
	if err != nil {
		return nil, err
	}

	recipes := make([]string, 0)
	for _, f := range recipeFiles {
		_, filename := path.Split(f)
		var extension = filepath.Ext(filename)
		if extension != ".json" {
			continue
		}
		var name = filename[0 : len(filename)-len(extension)]
		if runtime.GOOS == "windows" {
			// god I hate windows
			s := strings.Split(name, "\\")
			name = s[len(s)-1]
		}
		recipes = append(recipes, path.Base(name))
	}
	return recipes, nil
}

// recipesCmd
var recipesCmd = &cobra.Command{
	Use:   "recipes",
	Short: "List available environment recipes",
	Long:  `List available environment recipes`,
	Run: func(cmd *cobra.Command, args []string) {
		// get list of recipes from the home folder
		if CLIOptions.RecipesPath == "" {
			fmt.Println("recipes path not set")
			return
		}

		// get list of recipe files from the recipes folder
		recipes, err := GetRecipeList(CLIOptions.RecipesPath)
		if err != nil {
			fmt.Println("error getting recipe list")
			return
		}

		if CLIOptions.Json {
			// output as json
			output := make([]string, 0)
			output = append(output, recipes...)
			j, _ := util.ToJson(output, CLIOptions.PrettyJson)
			fmt.Println(j)
			return
		}

		// remove the path and extension from the recipe names
		for _, r := range recipes {
			fmt.Println(r)
		}
	},
}

func RecipeFromPath(path string) (*EnvRecipe, error) {
	recipe, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseRecipe(recipe)
}

// RecipePathFromName returns the full path to a recipe file given the recipe name
// the recipe name is the name of the file without the extension and must reside in the recipes folder
func RecipePathFromName(name string) (string, error) {
	if CLIOptions.RecipesPath == "" {
		return "", fmt.Errorf("recipes path not set")
	}
	retv := path.Join(CLIOptions.RecipesPath, name+".json")
	if _, err := os.Stat(retv); err != nil {
		return "", fmt.Errorf("recipe not found")
	}
	return retv, nil
}

func InitRecipes(envCmd *cobra.Command) {
	envCmd.AddCommand(recipesCmd)
}

// recipePathsFromNames is a recursive function that gets the full path to all recipes
func recipePathsFromNames(names []string, parsedrecipepaths *map[string]string) error {
	// get list of all recipe paths (including inherited recipes)
	for _, name := range names {
		if _, ok := (*parsedrecipepaths)[name]; ok {
			// already parsed this recipe paths
			continue
		}
		recipePath, err := RecipePathFromName(name)
		if err != nil {
			return err
		}
		(*parsedrecipepaths)[name] = recipePath
		// get the recipe
		r, err := RecipeFromPath(recipePath)
		if err != nil {
			return err
		}
		// get the inherited recipes
		if r.Inherits != nil {
			err = recipePathsFromNames(r.Inherits, parsedrecipepaths)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type daggerRecipeNode struct {
	dagger.IDaggerNode
	Recipe *EnvRecipe
}

func newDaggerRecipeNode(recipe *EnvRecipe) *daggerRecipeNode {
	node := &daggerRecipeNode{
		IDaggerNode: dagger.NewDaggerNode(),
		Recipe:      recipe,
	}

	node.SetName(recipe.Name)

	// add one autoclone input pin
	ipin := dagger.NewDaggerInputPin()
	ipin.SetAutoCloneMaster(ipin)
	ipin.SetMaxAutoClone(-1) // as many as we want
	node.GetInputPins(0).AddPin(ipin, "ip1")

	// add one output pin
	node.GetOutputPins(0).AddPin(dagger.NewDaggerOutputPin(), "op1")

	return node
}

// MergeRecipes merges two recipes into a new recipe
func MergeRecipes(r1 *EnvRecipe, r2 *EnvRecipe) (*EnvRecipe, error) {
	// we need to merge the two recipes into a new recipe
	// the new recipe will have the name of the first recipe

	// check if the recipes are compatible
	newname := ""
	if r1.Name == "" && r2.Name != "" {
		newname = r2.Name
	} else if r1.Name != "" && r2.Name == "" {
		newname = r1.Name
	} else {
		newname = r1.Name + "-" + r2.Name
	}

	newdescription := ""
	if r1.Description == "" && r2.Description != "" {
		newdescription = r2.Description
	} else if r1.Description != "" && r2.Description == "" {
		newdescription = r1.Description
	} else {
		newdescription = r1.Description + "\n" + r2.Description
	}

	retv := &EnvRecipe{
		Name:          newname,
		Description:   newdescription,
		RecipeFormat:  CurrentRecipeFormat,
		Version:       "1.0", // this will be a new recipe, so always 1.0
		PythonVersion: CurrentMinimumPython,
		Inherits:      make([]string, 0),
		ParamSets:     make(map[string][]string),
	}

	if r1.ParamSets != nil {
		retv.ParamSets = r1.ParamSets
	}
	if r2.ParamSets != nil {
		for k, v := range r2.ParamSets {
			if _, ok := retv.ParamSets[k]; ok {
				retv.ParamSets[k] = append(retv.ParamSets[k], v...)
			} else {
				retv.ParamSets[k] = v
			}
		}
	}

	// parse the python versions - take the highest version
	if r1.PythonVersion == "" {
		r1.PythonVersion = CurrentMinimumPython
	}
	if r2.PythonVersion == "" {
		r2.PythonVersion = CurrentMinimumPython
	}

	v1, err := kinda.ParseVersion(r1.PythonVersion)
	if err != nil {
		return nil, err
	}
	v2, err := kinda.ParseVersion(r2.PythonVersion)
	if err != nil {
		return nil, err
	}
	vdiff := v1.Compare(v2)
	if vdiff == 0 {
		// the versions are the same
		retv.PythonVersion = r1.PythonVersion
	} else if vdiff < 0 {
		// v2 is greater
		retv.PythonVersion = r2.PythonVersion
	} else {
		// v1 is greater
		retv.PythonVersion = r1.PythonVersion
	}

	// parse the channel
	if r1.Channel != nil && r2.Channel != nil {
		if *r1.Channel != *r2.Channel {
			slog.Warn(fmt.Sprintf("recipes %s and %s have different conda channels, using channel for %s", r1.Name, r2.Name, r2.Name))
			retv.Channel = r2.Channel
		}
		retv.Channel = r1.Channel
	} else if r1.Channel != nil {
		retv.Channel = r1.Channel
	} else if r2.Channel != nil {
		retv.Channel = r2.Channel
	}

	// parse the pip packages
	if r1.PipPackages != nil && r2.PipPackages != nil {
		// merge the pip packages
		retv.PipPackages = make([]PipPackages, 0)
		retv.PipPackages = append(retv.PipPackages, r1.PipPackages...)
		retv.PipPackages = append(retv.PipPackages, r2.PipPackages...)
	} else if r1.PipPackages != nil {
		retv.PipPackages = r1.PipPackages
	} else if r2.PipPackages != nil {
		retv.PipPackages = r2.PipPackages
	}

	// parse the custom nodes
	if r1.CustomNodes != nil || r2.CustomNodes != nil {
		// merge the custom nodes
		tcustomnodes := make(map[string]CustomNode)
		if r1.CustomNodes != nil {
			for _, c := range r1.CustomNodes {
				tcustomnodes[c.Name] = c
			}
		}
		if r2.CustomNodes != nil {
			for _, c := range r2.CustomNodes {
				tcustomnodes[c.Name] = c
			}
		}
		retv.CustomNodes = make([]CustomNode, 0)
		for _, c := range tcustomnodes {
			retv.CustomNodes = append(retv.CustomNodes, c)
		}
	}

	// parse the models
	if r1.Models != nil || r2.Models != nil {
		// merge the models
		tmodels := make(map[string]Models)
		if r1.Models != nil {
			for _, m := range r1.Models {
				tmodels[m.Name] = m
			}
		}
		if r2.Models != nil {
			for _, m := range r2.Models {
				tmodels[m.Name] = m
			}
		}
		retv.Models = make([]Models, 0)
		for _, m := range tmodels {
			retv.Models = append(retv.Models, m)
		}
	}

	return retv, nil
}

func recipeFromPaths(paths []string) (*EnvRecipe, error) {
	recipes := make(map[string]*EnvRecipe)
	for _, p := range paths {
		r, err := RecipeFromPath(p)
		if err != nil {
			return nil, err
		}
		recipes[r.Name] = r
	}

	// build a dependency graph
	graph := dagger.NewDaggerGraph(1)
	for _, r := range recipes {
		node := newDaggerRecipeNode(r)
		graph.AddNode(node, true)
	}

	// take each recipe name, get the node and connect it to the inherited recipes
	for _, r := range recipes {
		nodes := graph.GetNodesWithName(r.Name)
		if nodes == nil {
			return nil, fmt.Errorf("could not find recipe node %s", r.Name)
		} else if len(nodes) > 1 {
			return nil, fmt.Errorf("multiple nodes found for recipe %s", r.Name)
		}

		// get the output pin of the recipe node
		opin := nodes[0].GetOutputPins(0).GetPin("op1").(dagger.IDaggerOutputPin)
		if r.Inherits == nil {
			r.Inherits = make([]string, 0)
		}

		for _, i := range r.Inherits {
			inheritNodes := graph.GetNodesWithName(i)
			if inheritNodes == nil {
				return nil, fmt.Errorf("could not find inherited recipe %s", i)
			} else if len(inheritNodes) > 1 {
				return nil, fmt.Errorf("multiple nodes found for inherited recipe %s", i)
			}

			// get the first unconnected input pin of the inherited recipe node
			ipin := inheritNodes[0].GetFirstUnconnectedInputPin(0)

			if !opin.CanConnectToPin(ipin) {
				slog.Warn("cyclic dependency in recipe %s to inherited recipe %s", r.Name, i)
				continue
			}

			opin.ConnectToInput(ipin)
		}
	}

	graph.CalculateTopology()

	// there "should" be only one subgraph, we'll create an empty recipe and merge the recipe trees
	subgraphcount := graph.GetSubGraphCount(0)
	if subgraphcount != 1 {
		slog.Warn("recipe dependency graph has foundational recipes")
		emptyRecipe := &EnvRecipe{
			Inherits: make([]string, 0),
		}

		// get the nodes with the highest ordinal in each subgraph
		maxordinalnodes := make([]dagger.IDaggerNode, 0)
		for i := 0; i < subgraphcount; i++ {
			nodes := graph.GetSubGraphNodes(0, i)
			// find the max ordinal
			maxordinal := 0
			for _, v := range nodes {
				if v.GetOrdinal(0) > maxordinal {
					maxordinal = v.GetOrdinal(0)
				}
			}
			// get the nodes with the max ordinal
			for _, v := range nodes {
				if v.GetOrdinal(0) == maxordinal {
					maxordinalnodes = append(maxordinalnodes, v)
				}
			}
		}

		// create a node for the empty recipe and attach it to the max ordinal nodes
		emptyNode := newDaggerRecipeNode(emptyRecipe)
		graph.AddNode(emptyNode, false)
		for _, n := range maxordinalnodes {
			emptyNode.GetOutputPins(0).GetPin("op1").(dagger.IDaggerOutputPin).ConnectToInput(n.GetFirstUnconnectedInputPin(0))
		}
	}

	// get the nodes with the highest ordinal
	maxordinal := graph.GetMaxOrdinal(0)
	ordnodes := graph.GetNodesWithOrdinal(0, maxordinal)
	if len(ordnodes) != 1 {
		// same as above, we should only have one node, we'll create an empty recipe and merge the recipe trees
		emptyRecipe := &EnvRecipe{
			Inherits: make([]string, 0),
		}

		// create a node for the empty recipe and attach it to the max ordinal nodes
		emptyNode := newDaggerRecipeNode(emptyRecipe)
		graph.AddNode(emptyNode, true)
		for _, n := range ordnodes {
			emptyNode.GetOutputPins(0).GetPin("op1").(dagger.IDaggerOutputPin).ConnectToInput(n.GetFirstUnconnectedInputPin(0))
		}
		maxordinal = graph.GetMaxOrdinal(0)
	}

	// walk the orinals in reverse order and merge the recipes
	newrecipe := &EnvRecipe{
		RecipeFormat: CurrentRecipeFormat,
	}

	var err error
	for i := maxordinal; i >= 0; i-- {
		nodes := graph.GetNodesWithOrdinal(0, i)
		for _, n := range nodes {
			node := n.(*daggerRecipeNode)
			newrecipe, err = MergeRecipes(newrecipe, node.Recipe)
			if err != nil {
				return nil, err
			}
		}
	}

	return newrecipe, nil
}

func RecipeFromNames(names []string) (*EnvRecipe, error) {
	// get list of all recipe paths (including inherited recipes)
	// this will be a recursive operation
	recipesPaths := make(map[string]string)
	err := recipePathsFromNames(names, &recipesPaths)
	if err != nil {
		return nil, err
	}
	fmt.Println(recipesPaths)
	paths := make([]string, 0)
	for _, p := range recipesPaths {
		paths = append(paths, p)
	}
	return recipeFromPaths(paths)
}
