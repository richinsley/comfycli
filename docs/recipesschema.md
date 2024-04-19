# Comfycli Recipes JSON Schema
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ComfyCLI Recipe",
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "The name of the recipe."
    },
    "description": {
      "type": "string",
      "description": "A brief description of the recipe's purpose."
    },
    "version": {
      "type": "string",
      "description": "The version of the recipe."
    },
    "recipe_format": {
      "type": "string",
      "description": "The version of the recipe format.",
      "pattern": "^\\d+\\.\\d+$"
    },
    "inherits": {
      "type": "array",
      "description": "The base recipe(s) to inherit from.",
      "items": {
        "type": "string"
      }
    },
    "python": {
      "type": "string",
      "description": "The Python version to use."
    },
    "channel": {
      "type": "string",
      "description": "The conda channel to use."
    },
    "pip_packages": {
      "type": "array",
      "description": "The required pip packages.",
      "items": {
        "type": "object",
        "properties": {
          "extra_index_url": {
            "type": "string",
            "description": "The extra index URL for package installation."
          },
          "index_url": {
            "type": "string",
            "description": "The index URL for package installation."
          },
          "packages": {
            "type": "array",
            "description": "The list of pip packages to install.",
            "items": {
              "type": "object",
              "properties": {
                "name": {
                  "type": "string",
                  "description": "The name of the package."
                },
                "version": {
                  "type": "string",
                  "description": "The version of the package."
                }
              },
              "required": ["name"]
            }
          }
        },
        "required": ["packages"]
      }
    },
    "custom_nodes": {
      "type": "array",
      "description": "The custom node Git repositories to install.",
      "items": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string",
            "description": "The name of the custom node."
          },
          "git_url": {
            "type": "string",
            "description": "The Git URL of the custom node repository."
          },
          "branch": {
            "type": "string",
            "description": "The branch to clone from the repository."
          }
        },
        "required": ["name", "git_url"]
      }
    },
    "models": {
      "type": "array",
      "description": "The Stable Diffusion models to make available.",
      "items": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string",
            "description": "The name of the model."
          },
          "type": {
            "type": "string",
            "description": "The type of the model."
          },
          "base": {
            "type": "string",
            "description": "The base of the model."
          },
          "save_path": {
            "type": "string",
            "description": "The save path for the model."
          },
          "description": {
            "type": "string",
            "description": "A description of the model."
          },
          "reference": {
            "type": "string",
            "description": "A reference URL for the model."
          },
          "filename": {
            "type": "string",
            "description": "The filename of the model."
          },
          "url": {
            "type": "string",
            "description": "The URL to download the model from."
          }
        },
        "required": ["name", "type", "base", "save_path", "filename", "url"]
      }
    },
    "paramsets": {
      "type": "object",
      "description": "The parameter sets for the recipe.",
      "additionalProperties": {
        "type": "array",
        "items": {
          "type": "string"
        }
      }
    }
  },
  "required": ["name", "version", "recipe_format"]
}
```