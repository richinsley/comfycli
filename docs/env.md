# Environment Commands

The `env` command group in `comfycli` is designed to manage virtual Python environments for ComfyUI. These commands facilitate the creation, management, and deletion of environments, ensuring that each ComfyUI setup can be tailored and isolated according to specific needs.  It uses a [recipe](./recipes.md) based system that allows for constructing environments complete with specific sets of custom nodes and models.  Collections of recipes can be composed together to add features in a granular method.  

## Commands
- [pullrecipes](#pullrecipes): Import or update a recipe repo manifest from a URL
- [recipes](#recipes): List available recipes
- [setdefault](#setdefault): Set the default base recipe for the target system
- [create](#create): Create a new virtual ComfyUI python environment
- [ls](#ls): List available environments
- [runcomfy](#runcomfy): Launch ComfyUI within an environment
- [rm](#rm): Remove an environment
- [update](#update): Update an environment

***
## pullrecipes

**Description:** ***pullrecipes*** takes a URL to a manifest for a comfycli recipe repository.  Pulling a manifest into comfycli will import the recipes described in the repo and make them available for creating new environments.

**Usage:**
```bash
comfycli env pullrecipes [repo manifest URL] [flags]
```

**Examples:**
Pull the [primary recipe repository](https://github.com/richinsley/comfycli/tree/main/recipes):
```bash
:~$ comfycli env pullrecipes https://raw.githubusercontent.com/richinsley/comfycli/main/recipes/manifest.json
```

## recipes

**Description:** ***recipes*** lists the available recipes

**Usage:**
```bash
comfycli comfycli env recipes [flags]
```

**Examples:**
```bash
:~$ comfycli env recipes
SD15
SDXL
SDXL_lightning
controlnet_SD15
controlnet_SDXL
default
default_pytorch_nightly
default_pytorch_stable
ipadapter_v2
```

## setdefault

**Description:** ***setdefault*** sets the default base recipe for the target system.  Different archectures and GPUs may require different base recipes.  The base recipe is the root recipe that all other recipes will inherit from.  The first time any ***env*** command is run, if the default base recipe has not yet been set, the user will 
be asked to choose which recipe will be the default recipe.  ***setdefault*** allows the recipe to set at any time.  *Only recipes that are prefixed with "default_" can be set as default recipes.*

**Usage:**
```bash
comfycli comfycli env setdefault <recipe> [flags]
```

**Examples:**
Assuming we are running comfycli for the first time and try list the available recipes, we will be asked which recipe will be the default root recipe:
```bash
:~$ comfycli env recipes
Default recipe for the system is not set.
Please select a default recipe from the list below:
0: default_pytorch_nightly - Default macos environment for ComfyUI
1: default_pytorch_stable - Default macos environment for ComfyUI
Select one: 1
setting default recipe to default_pytorch_stable

SD15
SDXL
default
default_pytorch_nightly
default_pytorch_stable
```

At a later time, we can switch to a different default:
```bash
:~$ comfycli env setdefault default_pytorch_nightly
setting default recipe to default_pytorch_nightly
```