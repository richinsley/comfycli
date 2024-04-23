package pkg

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/richinsley/comfy2go/client"
	// "github.com/richinsley/comfy2go/graphapi"
	"github.com/schollz/progressbar/v3"
)

func ClientWithWorkflow(client_index int, options *ComfyOptions, workflowpath string, parameters []CLIParameter, callbacks *client.ComfyClientCallbacks) (*Workflow, bool, *[]string, error) {
	workflow, missing, err := GetFullWorkflow(client_index, options, workflowpath, callbacks)
	if missing != nil {
		return nil, false, missing, err
	}
	if err != nil {
		return nil, false, missing, err
	}

	hasPipeLoop, err := ApplyParameters(workflow.Client, options, workflow.Graph, workflow.SimpleAPI, parameters)
	if err != nil {
		return nil, false, nil, err
	}

	return workflow, hasPipeLoop, nil, nil
}

func ProcessQueue(options *ComfyOptions, workflowpath string, parameters []CLIParameter) bool {
	// callbacks can be used respond to QueuedItem updates, or client status changes
	callbacks := &client.ComfyClientCallbacks{
		ClientQueueCountChanged: func(c *client.ComfyClient, queuecount int) {
			slog.Debug(fmt.Sprintf("Client %s Queue size: %d", c.ClientID(), queuecount))
		},
		QueuedItemStarted: func(c *client.ComfyClient, qi *client.QueueItem) {
			slog.Debug(fmt.Sprintf("Queued item %s started", qi.PromptID))
		},
		QueuedItemStopped: func(cc *client.ComfyClient, qi *client.QueueItem, reason client.QueuedItemStoppedReason) {
			slog.Debug(fmt.Sprintf("Queued item %s stopped", qi.PromptID))
		},
		QueuedItemDataAvailable: func(cc *client.ComfyClient, qi *client.QueueItem, pmd *client.PromptMessageData) {
			slog.Debug(fmt.Sprintf("Queued item %s data available", qi.PromptID))
		},
	}

	workflow, hasPipeLoop, missing, err := ClientWithWorkflow(0, options, workflowpath, parameters, callbacks)
	if err != nil {
		slog.Error("Failed to create comfyui client", err)
		os.Exit(1)
	}

	// get any output nodes that were specified in the api
	var outputnodes map[string]bool = make(map[string]bool)
	if workflow.SimpleAPI != nil && workflow.SimpleAPI.OutputNodes != nil {
		for _, n := range workflow.SimpleAPI.OutputNodes {
			outputnodes[n.Title] = true
		}
	}

	// if CLIOptions.OutputNodes is set, we'll use that instead
	if options.OutputNodes != "" {
		outputnodes = make(map[string]bool)
		for _, n := range strings.Split(options.OutputNodes, ",") {
			outputnodes[n] = true
		}
	}

	outputnodeIDs := make(map[int]bool)
	for _, n := range workflow.Graph.Nodes {
		if _, ok := outputnodes[n.Title]; ok {
			outputnodeIDs[n.ID] = true
		}
	}

	if missing != nil {
		slog.Error("failed to get workflow: missing nodes", "missing", fmt.Sprintf("%v", missing))
		os.Exit(1)
	}
	if err != nil {
		slog.Error("Failed to create comfyui client", err)
		os.Exit(1)
	}

	item, err := workflow.Client.QueuePrompt(workflow.Graph)
	if err != nil {
		slog.Error("Failed to queue prompt", err)
		os.Exit(1)
	}

	// we'll provide a progress bar
	var bar *progressbar.ProgressBar = nil

	// continuously read messages from the QueuedItem until we get the "stopped" message type
	var currentNodeTitle string
	for continueLoop := true; continueLoop; {
		msg := <-item.Messages
		switch msg.Type {
		case "started":
			qm := msg.ToPromptMessageStarted()
			slog.Debug(fmt.Sprintf("Start executing prompt ID %s\n", qm.PromptID))
		case "executing":
			bar = nil
			qm := msg.ToPromptMessageExecuting()
			// store the node's title so we can use it in the progress bar
			currentNodeTitle = qm.Title
			slog.Debug(fmt.Sprintf("Executing Node: %d", qm.NodeID))
		case "progress":
			// update our progress bar
			qm := msg.ToPromptMessageProgress()
			if bar == nil {
				bar = progressbar.Default(int64(qm.Max), currentNodeTitle)
			}
			bar.Set(qm.Value)
		case "stopped":
			// if we were stopped for an exception, display the exception message
			qm := msg.ToPromptMessageStopped()
			if qm.Exception != nil {
				slog.Error(fmt.Sprintf("ComfyUI exception in node %s", qm.Exception.NodeName))
				slog.Error(qm.Exception.ExceptionMessage)
				os.Exit(1)
			}
			continueLoop = false
		case "data":
			qm := msg.ToPromptMessageData()
			// data objects have the fields: Filename, Subfolder, Type
			// * Subfolder is the subfolder in the output directory
			// * Type is the type of the image temp/
			for k, v := range qm.Data {
				// if qm.NodeID is not in outputnodeIDs, then we ignore the data
				if len(outputnodeIDs) != 0 {
					if _, ok := outputnodeIDs[qm.NodeID]; !ok {
						continue
					}
				}

				if k == "images" || k == "gifs" {
					for _, output := range v {
						img_data, err := workflow.Client.GetImage(output)
						if err != nil {
							slog.Error("Failed to get image", err)
							os.Exit(1)
						}

						// what to do with the image data
						if options.InlineImages {
							// print the image to the terminal
							OutputInlineToStd(img_data)
						}

						if !options.NoSaveData {
							SaveData(img_data, output.Filename)
						}

						if options.DataToStdout {
							_, err := os.Stdout.Write(*img_data)
							if err != nil {
								slog.Error("Failed to write data to stdout", err)
								os.Exit(1)
							}
							os.Stdout.Sync()
						}
						slog.Debug(fmt.Sprintf("Got data file: %s", output.Filename))
					}
				} else if k == "text" {
					for _, output := range v {
						fmt.Println(output.Text)
					}
				}
			}
		default:
			slog.Warn(fmt.Sprintf("Unknown message type: %s", msg.Type))
		}
	}

	// return true if we read from a pipe
	return hasPipeLoop
}
