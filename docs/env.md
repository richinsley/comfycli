# Environment Commands

The `env` command group in `comfycli` is designed to manage virtual Python environments for ComfyUI. These commands facilitate the creation, management, and deletion of environments, ensuring that each ComfyUI setup can be tailored and isolated according to specific needs.  It uses a [recipe](./recipes.md) based system that allows for constructing environments complete with specific sets of custom nodes and models.  Collections of recipes can be composed together to add features in a granular method.

## Commands
- [recipes](#recipes): List available recipes
- [setdefault](#setdefault): Set the default base recipe for the target system
- [create](#create): Create a new virtual ComfyUI python environment
- [ls](#ls): List available environments
- [runcomfy](#runcomfy): Launch ComfyUI within an environment
- [rm](#rm): Remove an environment
- [update](#update): Update an environment

***


