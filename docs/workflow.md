# Workflow Commands

## Commands
- [extract](#extract):Extract a workflow from PNG metadata
- [parse](#parse):Parse a workflow file and output the workflow json
- [api](#api):Output the API for the workflow in json format
- [queue](#queue):Queue a workflow for processing

## extract

**Description:** ***extract*** parses the workflow embedded in ComfyUI generated png files and outputs it to the stdout.  It can then be redirected to a file, or piped to other commands.

**Usage:**
```bash
comfycli workflow extract [png file path] [flags]
```

**Examples:**
```bash
# extrack the workflow from ComfyUI_00313_.png and save to output.json
comfycli workflow extract ComfyUI_00313_.png > output.json
```