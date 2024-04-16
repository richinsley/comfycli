# Environment Commands

The `env` command group in `comfycli` is designed to manage virtual Python environments for ComfyUI. These commands facilitate the creation, management, and deletion of environments, ensuring that each ComfyUI setup can be tailored and isolated according to specific needs.  It uses a [recipe](./recipes.md) based system that allows for constructing environments complete with specific sets of custom nodes and models.  Collections of recipes can be composed together to add features in a granular method.  

## Commands
- [pullrecipes](#pullrecipes): Import or update a recipe repo manifest from a URL
- [recipes](#recipes): List available recipes
- [setdefault](#setdefault): Set the default base recipe for the target system
- [create](#create): Create a new virtual ComfyUI python environment
- [ls](#ls): List available environments
- [runcomfy](#runcomfy): Launch ComfyUI within an environment
- [update](#update): Update an environment
- [rm](#rm): Remove an environment

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

## create

**Description:** ***create*** a ComfyUI virtual environment.  The create command utilizes [micromamba](https://mamba.readthedocs.io/en/latest/user_guide/micromamba.html) to generate a python virtual environment based off of a given [recipe](./recipes.md) or a collection of recipes.  Comfycli recipes describe the environment, which version of python to use, any required pip packages for a given instance, custom nodes, and models.  This allows us to achieve consistent environments that are easy to deploy and redeploy for any given task. Comfycli environments can then be launched with the [runcomfy](#runcomfy) command.



**Flags:**
```
    --file string     Path to an external recipe file to use for environment creation
-h, --help            help for create
    --name string     Name of the environment to create (default "default")
-n, --noshared        Do not use shared models path
    --python string   Override the recipe python version to use
-q, --quiet           Silent output
    --recipe string   Recipe to use for environment creation (default "default")
    --verbose         Verbose output
```

**Usage:**
```bash
comfycli env create [flags]
```

**Examples:**
```bash
# create a new environment using the default system recipe with the name 'default'
comfycli env create

# create a new environment using the default system recipe with the name 'myenv'
comfycli env create --name myenv

# create a new environment using the default recipe, force it to be python 3.10 and name it 'myenv311'
comfycli env create --recipe default --python 3.11 --name myenv311

# combine multiple recipes (SD15 and SDXL) into a new environment
comfycli env create --recipe default,SD15,SDXL --name all_sd
```

## ls

**Description:** ***ls*** lists the current created environments

**Usage:**
```bash
comfycli env ls [flags]
```

**Examples:**

```bash
:~$ comfycli env ls
all_sd
default
myenv
myenv311
```

## runcomfy

**Description:** ***runcomfy*** launches an instance ComfyUI within a given virtual environment environment.  To pass ComfyUI cli parameters to the ComfyUI instance, place them after a "--" parameter separator.

**Usage:**
```bash
comfycli env runcomfy <environment> [flags]
```

**Examples:**

```bash
# run ComfyUI in the default environment
:~$ comfycli env runcomfy

# run ComfyUI in the myenv environment.  Pass the --listen and --highvram arguments to the ComfyUI script (notice the "--" that separates comfycli parameters from the comfyui parameters)
:~$ comfycli env runcomfy myenv -- --listen --highvram
```

## update

**Description:** ***update*** will update the ComfyUI git repo and all it's custom node git repos to the current versions.

**Usage:**
```bash
comfycli env update <environment> [flags]
```

**Examples:**

```bash
:~$ comfycli env update myenv
Update environment myenv: /Users/richardinsley/.comfycli/environments/envs/myenv [Y]: 
Updating environment myenv: /Users/richardinsley/.comfycli/environments/envs/myenv
Updating ComfyUI repository
ComfyUI repository already up-to-date
Updating custom node repository: ComfyUI-Manager
Custom node repository ComfyUI-Manager already up-to-date
```

## rm

**Description:** ***rm*** removes an environment from the comfycli home folder.

**Usage:**
```bash
comfycli env rm <environment> [flags]
```

**Examples:**

```bash
# remove the environment "myenv311"
:~$ comfycli env rm myenv311
Are you sure you want to remove environment myenv311: /Users/richardinsley/.comfycli/environments/envs/myenv311 [Y]: 
Removing environement myenv311: /Users/richardinsley/.comfycli/environments/envs/myenv311
```