# Comfycli Recipes

ComfyCLI is a command-line interface application that allows for the creation of virtual environments to run ComfyUI. It uses a recipe-based system to describe various aspects of the environment, such as the Python version, required pip packages, custom node Git repositories, and models.

## Recipe Concepts
- Default Recipe: The default recipe serves as the base recipe from which all other recipes inherit. It is tailored to the specific requirements of the target system (Linux, macOS, or Windows).
    - Default recipes are named with the prefix default_ (e.g., default_cuda.json, default_directml.json).
    - When ComfyCLI is run for the first time, the user selects the appropriate default recipe based on their system's hardware and requirements.
    - The selected default recipe is copied to a file named default.json.
- Recipe Inheritance: Recipes can inherit from one or more recipes, allowing for modular and reusable configurations.
    - Inherited recipes include all the settings and dependencies from their base recipes.
    - Recipes can override or extend the settings and dependencies inherited from their base recipes.
- Directed Acyclic Graph (DAG): ComfyCLI uses a DAG to resolve dependencies when combining multiple inherited recipes to create a new environment.
    - The DAG ensures that dependencies are resolved in the correct order and avoids circular dependencies.

## Creating Recipes

To create a new recipe, follow these steps:
1. Create a new JSON file with a descriptive name (e.g., **my_recipe.json**).
2. Define the recipe's metadata:
    * ***name***: The name of the recipe.
    * ***description***: A brief description of the recipe's purpose.
    * ***version***: The version of the recipe.
    * ***recipe_format***: The version of the recipe format (e.g., "1.0").
3. Specify the base recipe(s) to inherit from using the ***inherits*** field (optional).
4. Define the Python version to use with the ***python*** field (optional, defaults to the base recipe's Python version).
5. Specify the required pip packages using the ***pip_packages*** field (optional).
6. List the custom node Git repositories to install using the ***custom_nodes*** field (optional).
7. Specify the  models to make available using the ***models*** field (optional).  The models field follows the same format of [ComfyUI-Manager](https://github.com/ltdrdata/ComfyUI-Manager/blob/main/model-list.json) model list.

#### An example recipe from the [comfycli-sample-recipes](https://github.com/richinsley/comfycli-sample-recipes) repo that installs the SDXL Lightning Lora:
```json
{
    "name": "SDXL_lightning",
    "description": "SDXL Lightning Lora",
    "reference": "https://huggingface.co/ByteDance/SDXL-Lightning",
    "version": "1.0",
    "recipe_format": "1.0",
    "inherits": ["SDXL"],
    "models": [
        {
            "name": "SDXL Lightning LoRA (2step)",
            "type": "lora",
            "base": "SDXL",
            "save_path": "loras/SDXL-Lightning",
            "description": "SDXL Lightning LoRA (2step)",
            "reference": "https://huggingface.co/ByteDance/SDXL-Lightning",
            "filename": "sdxl_lightning_2step_lora.safetensors",
            "url": "https://huggingface.co/ByteDance/SDXL-Lightning/resolve/main/sdxl_lightning_2step_lora.safetensors"
          },
          {
            "name": "SDXL Lightning LoRA (4step)",
            "type": "lora",
            "base": "SDXL",
            "save_path": "loras/SDXL-Lightning",
            "description": "SDXL Lightning LoRA (4step)",
            "reference": "https://huggingface.co/ByteDance/SDXL-Lightning",
            "filename": "sdxl_lightning_4step_lora.safetensors",
            "url": "https://huggingface.co/ByteDance/SDXL-Lightning/resolve/main/sdxl_lightning_4step_lora.safetensors"
          },
          {
            "name": "SDXL Lightning LoRA (8step)",
            "type": "lora",
            "base": "SDXL",
            "save_path": "loras/SDXL-Lightning",
            "description": "SDXL Lightning LoRA (8tep)",
            "reference": "https://huggingface.co/ByteDance/SDXL-Lightning",
            "filename": "sdxl_lightning_8step_lora.safetensors",
            "url": "https://huggingface.co/ByteDance/SDXL-Lightning/resolve/main/sdxl_lightning_8step_lora.safetensors"
          }
    ]
}
```

## Recipe Format ([JSON Schema here](./recipesschema.md))
The recipe format follows a specific structure:

* ***name*** (string): The name of the recipe.
* ***description*** (string, optional): A brief description of the recipe's purpose.
* ***version*** (string): The version of the recipe.
* ***recipe_format*** (string): The version of the recipe format (e.g., "1.0").
* ***inherits*** (array of strings, optional): The base recipe(s) to inherit from.
* ***python*** (string, optional): The Python version to use (defaults to the base recipe's Python version).
* ***channel*** (string, optional): The conda channel to use (e.g., "conda-forge").
* ***pip_packages*** (array of objects, optional): The required pip packages.
    * ***extra_index_url*** (string, optional): The extra index URL for package installation.
    * ***index_url*** (string, optional): The index URL for package installation.
    * ***packages*** (array of objects): The list of pip packages to install.
        * ***name*** (string): The name of the package.
        * ***version*** (string, optional): The version of the package.
* ***custom_nodes*** (array of objects, optional): The custom node Git repositories to install.
    * ***name*** (string): The name of the custom node.
    * ***git_url*** (string): The Git URL of the custom node repository.
    * ***branch*** (string, optional): The branch to clone from the repository.
* ***models*** (array of objects, optional): The Stable Diffusion models to make available.
    * ***name*** (string): The name of the model.
    * ***type*** (string): The type of the model (e.g., "checkpoints").
    * ***base*** (string): The base of the model.
    * ***save_path*** (string): The save path for the model.
    * ***description*** (string, optional): A description of the model.
    * ***reference*** (string, optional): A reference URL for the model.
    * ***filename*** (string): The filename of the model.
    * ***url*** (string): The URL to download the model from.