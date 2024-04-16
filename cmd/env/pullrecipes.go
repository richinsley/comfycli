/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package env

import (
	"encoding/json"
	"fmt"
	"io"

	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	util "github.com/richinsley/comfycli/pkg"
	kinda "github.com/richinsley/kinda/pkg"
	"github.com/spf13/cobra"
)

type RecipeManifestEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	// Description is optional
	Description string `json:"description,omitempty"`
}

type RecipeRepoManifest struct {
	RepoName string                `json:"name"`
	Recipes  []RecipeManifestEntry `json:"recipes"`
	Origin   string                `json:"origin,omitempty"`
}

// fetchData decides whether to fetch data from HTTP/HTTPS or from a local file
func fetchData(u string) (string, error) {
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return fetchHTTP(u)
	} else {
		return readLocalFile(strings.TrimPrefix(u, "file://"))
	}
}

// fetchHTTP handles fetching data from HTTP and HTTPS URLs
func fetchHTTP(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// readLocalFile handles reading data from a local file
func readLocalFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// resolvePath combines base URL or directory with a relative file path
func resolvePath(base *url.URL, relativePath string) string {
	if base.Scheme == "http" || base.Scheme == "https" {
		// Create a new URL by resolving the relative path against the base URL
		newURL, err := base.Parse(relativePath)
		if err != nil {
			fmt.Println("Error resolving URL:", err)
			return ""
		}
		return newURL.String()
	}
	// For file paths, use filepath to resolve it
	return filepath.Join(filepath.Dir(base.Path), relativePath)
}

func toURLString(path string) (string, error) {
	// Check if the path already appears to be a URL
	if strings.Contains(path, "://") {
		return path, nil
	}
	// Assume it's a local file path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	// Convert local path to file URL
	return "file://" + absPath, nil
}

func PullRecipeRepoManifest(manifesturl string) (*RecipeRepoManifest, error) {
	jsonData, err := fetchData(manifesturl)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		os.Exit(1)
	}

	// Parse the JSON data
	var manifest RecipeRepoManifest
	err = json.Unmarshal([]byte(jsonData), &manifest)
	if err != nil {
		fmt.Println("Error parsing JSON data:", err)
		os.Exit(1)
	}

	manifest.Origin, err = toURLString(manifesturl)
	if err != nil {
		fmt.Println("Error converting to URL:", err)
		os.Exit(1)
	}

	return &manifest, nil
}

// pullrecipesCmd
var pullrecipesCmd = &cobra.Command{
	Use:   "pullrecipes",
	Short: "Import or update a recipe collection manifest from a URL",
	Long:  `Import or update a recipe collection manifest from a URL`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			slog.Error("error: no recipe manifest URL specified")
			os.Exit(1)
		}
		manifesturl := args[0]

		manifest, err := PullRecipeRepoManifest(manifesturl)
		if err != nil {
			slog.Error("error: %v", err)
			os.Exit(1)
		}

		// Determine base URL or directory
		base, err := url.Parse(manifest.Origin)
		if err != nil {
			fmt.Println("Error parsing base URL:", err)
			os.Exit(1)
		}

		// check to see if this repository has already been imported to recipes/repos
		repofilepath := filepath.Join(CLIOptions.RecipesRepos, manifest.RepoName+".json")
		if _, err := os.Stat(repofilepath); err == nil {
			// decode the existing file
			repofile, err := os.Open(repofilepath)
			if err != nil {
				slog.Warn("error opening existing repo file: %v", err)
			} else {
				repodata, err := io.ReadAll(repofile)
				if err != nil {
					slog.Warn("error reading existing repo file: %v", err)
				} else {
					var existingRepo RecipeRepoManifest
					err = json.Unmarshal(repodata, &existingRepo)
					if err != nil {
						slog.Warn("error parsing existing repo file: %v", err)
					} else {
						// check if they are from the same origin
						if existingRepo.Origin == manifest.Origin {
							// take each entry in the incoming manifest and compare it to the existing manifest
							for _, entry := range manifest.Recipes {
								// check if the entry is in the existing manifest
								index := -1
								for i, existingentry := range existingRepo.Recipes {
									if entry.Name == existingentry.Name {
										index = i
										break
									}
								}
								if index != -1 {
									// if the version is newer, replace the existing entry
									existingRecipeVersion, _ := kinda.ParseVersion(existingRepo.Recipes[index].Version)
									incomingRecipeVersion, _ := kinda.ParseVersion(entry.Version)
									if incomingRecipeVersion.Compare(existingRecipeVersion) == 1 {
										existingRepo.Recipes[index] = entry
									}
								} else {
									// if the entry is not in the existing manifest, add it
									existingRepo.Recipes = append(existingRepo.Recipes, entry)
								}
							}

							// write the updated manifest back to the file
							repofile, err := os.Create(repofilepath)
							if err != nil {
								slog.Warn("error creating repo file: %v", err)
							} else {
								repodata, err := json.Marshal(existingRepo)
								if err != nil {
									slog.Warn("error marshalling repo data: %v", err)
								} else {
									_, err = repofile.Write(repodata)
									if err != nil {
										slog.Warn("error writing repo data: %v", err)
									}
								}
							}
						} else {
							// if they are not from the same origin, prompt the user to overwrite the existing repo
							slog.Warn(fmt.Sprintf("error: repository %s has already been imported from a different origin", manifest.RepoName))
							proceed, err := util.YesNo("Overwrite existing repo?", false)
							if err != nil {
								slog.Error("invalid response", "error", err)
								os.Exit(1)
							}
							if !proceed {
								os.Exit(0)
							}

							// write the new manifest to the file
							repofile, err := os.Create(repofilepath)
							if err != nil {
								slog.Warn("error creating repo file: %v", err)
							} else {
								repodata, err := json.Marshal(manifest)
								if err != nil {
									slog.Warn("error marshalling repo data: %v", err)
								} else {
									_, err = repofile.Write(repodata)
									if err != nil {
										slog.Warn("error writing repo data: %v", err)
									}
								}
							}
						}
					}
				}
			}
		} else {
			// write the new manifest to the file
			repofile, err := os.Create(repofilepath)
			if err != nil {
				slog.Warn("error creating repo file: %v", err)
			} else {
				repodata, err := json.Marshal(manifest)
				if err != nil {
					slog.Warn("error marshalling repo data: %v", err)
				} else {
					_, err = repofile.Write(repodata)
					if err != nil {
						slog.Warn("error writing repo data: %v", err)
					}
				}
			}
		}

		// if we're here, we have a valid manifest and a valid file path where it was saved
		// open and decode the manifest file
		repofile, err := os.Open(repofilepath)
		if err != nil {
			slog.Error("error opening repo file: %v", err)
			os.Exit(1)
		}
		repodata, err := io.ReadAll(repofile)
		if err != nil {
			slog.Error("error reading repo file: %v", err)
			os.Exit(1)
		}
		var repo RecipeRepoManifest
		err = json.Unmarshal(repodata, &repo)
		if err != nil {
			slog.Error("error parsing repo file: %v", err)
			os.Exit(1)
		}

		// Fetch or read each file listed in the JSON
		for _, file := range repo.Recipes {
			fullPath := resolvePath(base, file.Name+".json")
			fileData, err := fetchData(fullPath)
			if err != nil {
				fmt.Println("Error fetching file", file, ":", err)
				continue
			}

			// parse the JSON data into a EnvRecipe
			recipe, err := ParseRecipe([]byte(fileData))
			if err != nil {
				fmt.Println("Error parsing recipe", file, ":", err)
				continue
			}

			// see if the recipe already exists
			recipePath := filepath.Join(CLIOptions.RecipesPath, recipe.Name+".json")
			if _, err := os.Stat(recipePath); err == nil {
				// compare the versions
				var existingRecipe EnvRecipe
				existingRecipeData, err := os.ReadFile(recipePath)
				if err != nil {
					fmt.Println("Error reading existing recipe", recipe.Name, ":", err)
					recipe.WriteRecipe(recipePath, true)
					continue
				}
				err = json.Unmarshal(existingRecipeData, &existingRecipe)
				if err != nil {
					fmt.Println("Error parsing existing recipe", recipe.Name, ":", err)
					recipe.WriteRecipe(recipePath, true)
					continue
				}
				existingVersion, err := kinda.ParseVersion(existingRecipe.Version)
				if err != nil {
					fmt.Println("Error parsing existing recipe version", recipe.Name, ":", err)
					recipe.WriteRecipe(recipePath, true)
					continue
				}
				incomingVersion, err := kinda.ParseVersion(recipe.Version)
				if err != nil {
					fmt.Println("Error parsing incoming recipe version", recipe.Name, ":", err)
					continue
				}
				if incomingVersion.Compare(existingVersion) == 1 {
					// overwrite the existing recipe
					err = recipe.WriteRecipe(recipePath, true)
					if err != nil {
						fmt.Println("Error writing updated recipe", recipe.Name, ":", err)
						continue
					}
				}
			} else {
				// write the new recipe
				err = recipe.WriteRecipe(recipePath, false)
				if err != nil {
					fmt.Println("Error writing new recipe", recipe.Name, ":", err)
					continue
				}
				fmt.Printf("Imported recipe %s\t\t%s\n", recipe.Name, recipe.Description)
			}
		}
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		CheckForDefaultRecipe()
	},
}

func InitPullRecipes(envCmd *cobra.Command) {
	envCmd.AddCommand(pullrecipesCmd)
}
