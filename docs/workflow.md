# Workflow Commands

## Commands
- [extract](#extract):Extract a workflow from PNG metadata
- [inject](#extract):Inject a workflow into PNG metadata
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

## inject

**Description:** ***inject*** loads a workflow from json or a png file, then injects it into a new png file.  This allows for scenarios such as taking a screen show of a workflow from the ComfyUI interface, and injecting the actual workflow into the screen shot metadata.
If no output file is specified with "--output", then the png is written to stdout.

**Usage:**
```bash
comfycli workflow inject [workflow.json] [png file path] [flags]
```

**Flags:**
```bash
-o, --output string   Path to write new PNG file with metadata
```

**Examples:**
```bash
# Load a workflow from a png file, inject it into a PNG file and write the new PNG to stdout
comfycli workflow inject workflow.png image.png > newimage.png

# Load a workflow from a json file, inject it into a PNG file and write the new PNG to newimage.png
comfycli workflow inject workflow.json image.png --output newimage.png
```

## parse

**Description:** ***parse*** a workflow file and output the workflow json.  This allows you to modify the parameters of a workflow and output the json
to a file or the terminal.  Parameters follow the delimiter "--" and are in the format "node:parameter"=value.  For example, to set the seed parameter of a KSampler node to 1234, you would use "KSampler:seed"=1234

**Usage:**
```bash
comfycli workflow parse [workflow file] [flags]
```

**Flags:**
```bash
-g, --graphout string   Path to write workflow graph JSON
```

**Examples:**
```bash
# parse the default workflow, set the KSampler seed parameter to 1234 and output the workflow json to a file
comfycli workflow parse defaultworkflow.json -- "KSampler:seed"=1234 > newworkflow.json
```

## queue

**Description:** ***queue*** a workflow for processing. The first argument is the path to the workflow file.  Set the parameters for the workflow by adding them as additional arguments after "--"
Node parameters are set by providing the node name followed by the parameter name and value.
When using a [Simple API](./simpleapi.md), parameters can be set by providing the just the parameter name and value. Nodes that output data save the data to the current working directory unless the "--nosavedata" flag is set.  When an output node is defined in a Simple API, only those output nodes save data.

Comfycli supports displaying the output images in the terminal by leveraging the iTerm2 [Inline Images Protocol](https://iterm2.com/documentation-images.html).
Some supported terminal emulators:
* [iTerm2](https://iterm2.com/index.html) (macOS)
* [Wezterm](https://wezfurlong.org/wezterm/index.html) (Windows, macOS, Linux)
* [konsole](https://konsole.kde.org/) (Linux)
* [mintty](https://mintty.github.io/) (Windows)

**Usage:**
```bash
comfycli workflow parse [workflow file] [flags]
```

**Flags:**
```bash
  -i, --inlineimages         Output images to terminal with Inline Image Protocol
  -n, --nosavedata           Do not save data to disk
  -o, --outputnodes string   Specify which output nodes save data. Comma separated nodes. (Default is all nodes)
```

**Examples:**
```bash
# Set the seed parameter for a node with the title "KSampler"
comfycli workflow queue myworkflow.json -- KSampler:seed=1234

# Use a workflow that has a Simple API that has a parameter named "seed"
comfycli --api API workflow queue myworkflow_simple_api.json -- seed=1234

# Queue a workflow, don't save images to disk, but output them to the terminal using the Inline Image Protocol
comfycli workflow queue --inlineimages --nosavedata myworkflow.json -- KSampler:seed=1234
```