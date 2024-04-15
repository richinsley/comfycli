# System Commands

The `system` command group in `comfycli` provides tools to manage and interact with your ComfyUI instance at a system level. These commands allow you to check system capabilities, view configurations, and monitor real-time operations.  When using the system commands, the target ComfyUI host can be specified with the "--host" flag or with the COMFYCLI_HOST environment variable.

## Commands
- [canrun](#canrun): Tests if a ComfyUI instance can run a specified workflow.
- [nodes](#nodes): List all available nodes.
- [info](#info): Retrieve detailed system information.
- [top](#top): Provides a real-time view of system information.
- [wait](#wait): Waits for the job queue to be empty.

***
## canrun

**Description:** Tests if a ComfyUI instance can run a specified workflow by checking system resources and required nodes.  Canrun checks to see if the required nodes are installed in the target ComfyUI instance, and will check if all combo box values are valid.  The output can be plain text, or json when using the "-j" flag.

**Usage:**
```bash
comfycli system canrun [workflow file path] [flags]
```

**Examples:**
$Check if ComfyUI running on the local host can run the workflow SDXL.json.  This particular instance is missing a model that is required:
```bash
:~$ comfycli system canrun SDXL.json
failed to get workflow
missing combo values:
--------------
{Load Checkpoint CheckpointLoaderSimple ckpt_name thinkdiffusionxl_v10.safetensors}
```

By providing the "-j" flag, we get json output which specifies which nodes or models are missing:
```bash
:~$ comfycli system canrun -j SDXL.json
{
    "canrun": false,
    "missing_combo_values": [
        {
            "NodeTitle": "Load Checkpoint",
            "NodeType": "CheckpointLoaderSimple",
            "PropertyName": "ckpt_name",
            "PropertyValue": "thinkdiffusionxl_v10.safetensors"
        }
    ]
}
```
## nodes

**Description:** List available nodes in a ComfyUI instance.  The nodes command will output all the availables nodes in the target ComfyUI instance along with ech node's available properties.  By providing the "-j" flag, it will output in json format.

**Usage:**
```bash
comfycli system nodes [flags]
```

## info

**Description:** Retrieve system information from a ComfyUI instance.

**Usage:**
```bash
comfycli system info [flags]
```

**Examples:**
Get system info for a ComfyUI instance running on the host 192.168.0.51:8188
```bash
:~$ comfycli system info --host 192.168.0.51:8188
System Info:
  OS: posix
  Python Version: 3.10.14 | packaged by conda-forge | (main, Mar 20 2024, 12:51:49) [Clang 16.0.6 ]
  Embedded Python: false
Devices:
  Name: mps
  Type: mps
  Index: 0
  VRAM Total: 103079215104
  VRAM Free: 65021968384
  Torch VRAM Total: 103079215104
  Torch VRAM Free: 65021968384
```

## top

**Description:** Continuously retrieveand display system information from a ComfyUI instance.

**Flags:**
-i, --interval int   Interval in seconds to update the top information (default 4)

**Usage:**
```bash
comfycli system info [flags]
```

## wait

**Description:** Wait for a ComfyUI instance's queue to empty.  The wait command will block until the queue count reaches 0.


**Usage:**
```bash
comfycli system wait [flags]
```