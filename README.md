# Comfycli


## About
Comfycli is a command-line interface designed to enhance the user experience of ComfyUI by providing powerful scripting and automation capabilities directly from the command line. Tailored for developers and users keen on stable diffusion models, comfycli simplifies the management of intricate AI workflows and supports a recipe-based system for creating and managing virtual environments. This feature allows users to define and replicate environments with precision, ensuring consistent setups across different machines or projects.
Comfycli aims to bridge the gap between graphical interface usability and command-line efficiency, allowing for more precise control over the configurations and operations of ComfyUI environments.
## Installation

Installation scripts are provided for quick and easy installation via the command line.

For Linux, macOS, or Git Bash on Windows install with:
```bash
"${SHELL}" <(curl -L https://raw.githubusercontent.com/richinsley/comfycli/main/install_scripts/installer.sh)
```

For Windows Powershell:
```powershell
Invoke-Expression ((Invoke-WebRequest -Uri https://raw.githubusercontent.com/richinsley/comfycli/main/install_scripts/installer.ps1).Content)
```

Install via Go:
```bash
go install github.com/richinsley/comfycli
```
## Usage
Comfycli is built on a structure of commands/subcommands.  The first time comfycli is run, it will set up a home directory:
```bash
❯ ./comfycli --help
Creating comfycli home folder: /Users/richardinsley/comfycli
Creating models folder: /Users/richardinsley/comfycli/models
Creating recipes folder: /Users/richardinsley/comfycli/environments/recipes
A feature-rich command-line application designed to streamline 
the interaction with and scripting for ComfyUI for a shell.

Version: 0.0.1
Home Path: /Users/richardinsley/comfycli

Usage:
  comfycli [flags]
  comfycli [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  env         Create and manage python virtual environments for ComfyUI
  help        Help about any command
  system      System commands for a ComfyUI instance
  workflow    Perform workflow operations with a ComfyUI instance

Flags:
      --api string         Simple API title (default "API")
      --apivalues string   Path to API values JSON or '-' for stdin
  -g, --graphout string    Path to write workflow graph JSON
  -h, --help               help for comfycli
      --host string        Host address (default "127.0.0.1:8188")
  -j, --json               Report all output as json
  -s, --stdout             Write node output data to stdout
  -v, --version            Print the version of comfycli
  -y, --yes                Automatically answer yes on prompted questions

Use "comfycli [command] --help" for more information about a command.
```

## Contributing

Pull requests are welcome for bug fixes and new features. For major changes, please open an issue first
to discuss what you would like to change.


## License

[MIT](https://choosealicense.com/licenses/mit/)