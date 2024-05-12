/*
Copyright © 2024 Rich Insley <richinsley@gmail.com>
*/
package workflow

import (
	"fmt"
	"os"
	"time"

	"github.com/richinsley/comfycli/pkg"
	"github.com/spf13/cobra"
)

// optional file server
var filesrv *pkg.FileServer = nil

// queueCmd represents the queue command
var queueCmd = &cobra.Command{
	Use:   "queue [workflow file]",
	Short: "Queue a workflow for processing",
	Long: `
Queue a workflow for processing. The first argument is the path to the workflow file.
Set the parameters for the workflow by adding them as additional arguments after "--"
Node parameters are set by providing the node name followed by the parameter name and value.
When using a Simple API, parameters can be set by providing the parameter name and value.
Nodes that output data save the data to the current working directory.  
An optional file server can be started to serve the files by specifying "--servepath".  
For more robust file serving, use the "util fileserve" command.

examples:

# Set the seed parameter for a node with the title "KSampler"
comfycli workflow queue myworkflow.json -- KSampler:seed=1234

# Use a workflow that has a Simple API that has a parameter named "seed"
comfycli --api API workflow queue myworkflow_simple_api.json -- seed=1234

# Queue a workflow, don't save images to disk, but output them to the terminal using the Inline Image Protocol
comfycli workflow queue --inlineimages --nosavedata myworkflow.json -- KSampler:seed=1234

# Queue a workflow, and open a file server to serve files
comfycli workflow queue myworkflow.json --serveport 8080 --servepath /path/to/files
`,
	Run: func(cmd *cobra.Command, args []string) {
		workflowPath := args[0]
		params := args[1:] // All other args are considered parameters
		parameters := pkg.ParseParameters(params)

		hasloop, err := pkg.TestParametersHasPipeLoop(CLIOptions, parameters)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// do we need to enable the file server?
		servePort, _ := cmd.Flags().GetInt("serveport")
		servePath, _ := cmd.Flags().GetString("servepath")
		if servePath != "" {
			options := pkg.FileServerOptions{
				Port:        servePort,
				RootPath:    servePath,
				StoragePath: servePath,
			}

			// create the file server
			filesrv, err = pkg.StartFileServer(options)
			if err != nil {
				fmt.Printf("error starting file server: %v\n", err.Error())
				os.Exit(1)
			}
		}

		if hasloop && len(CLIOptions.Host) > 1 {
			// get the workflows for each host that can process the workflow
			// the workers channel is filled asynchronously as the workflows are created
			tmpworkers := pkg.GetWorkflowsAsync(CLIOptions, workflowPath, parameters)

			// create a channel to hold the workers
			workers := make(chan *pkg.WorkflowQueueProcessor, len(CLIOptions.Host))

			// fill the workers channel and check for errors
			workercount := 0
			for i := 0; i < len(CLIOptions.Host); i++ {
				w := <-tmpworkers
				// try to cast to a WorkflowQueueProcessor
				if wqp, ok := w.(*pkg.WorkflowQueueProcessor); ok {
					workers <- wqp
					workercount++
				} else {
					// cast to error
					if err, ok := w.(error); ok {
						fmt.Printf("Error creating workflow client: %v\n", err.Error())
					}
				}
			}

			if workercount == 0 {
				fmt.Println("No client could be created to process the workflow")
				os.Exit(1)
			} else if workercount == 1 {
				// Process the queue on a single client. If there was a pipe loop, process it again
				for {
					hasPipeLoop, err := pkg.ProcessQueue(CLIOptions, workflowPath, parameters)
					if err != nil && hasloop {
						// not an actual error, just ran out of parameter inputs
						break
					}

					if !hasPipeLoop {
						break
					}
				}
			} else {
				// should the results be ordered?
				ordered, _ := cmd.Flags().GetBool("ordered")
				batchQueueProcess(workercount, workers, parameters, ordered)
			}
		} else {
			// Process the queue on a single client. If there was a pipe loop, process it again
			for {
				hasPipeLoop, err := pkg.ProcessQueue(CLIOptions, workflowPath, parameters)
				if err != nil && hasloop {
					// not an actual error, just ran out of parameter inputs
					break
				}
				if !hasPipeLoop {
					break
				}
			}
		}

		if filesrv != nil {
			pkg.StopFileServer(filesrv)
		}
	},
}

func batchQueueProcess(workercount int, workers chan *pkg.WorkflowQueueProcessor, parameters []pkg.CLIParameter, ordered bool) {
	deadcount := 0
	workitem := 0
	var dataitems chan pkg.WorkflowQueueDataOutputItems = nil
	if ordered {
		dataitems = make(chan pkg.WorkflowQueueDataOutputItems, workercount*3)
	}

	// Create a map to store the received data items
	receivedItems := make(map[int]pkg.WorkflowQueueDataOutputItems)
	nextExpectedItem := 0

	// Process the received data items concurrently
	go func() {
		for item := range dataitems {
			receivedItems[item.WorkItem] = item

			// Process the items in order
			for {
				if item, ok := receivedItems[nextExpectedItem]; ok {
					// Process the items
					for _, v := range item.Outputs {
						pkg.HandleDataOutput(item.Client, CLIOptions, v)
					}

					// Remove the processed item from the map
					delete(receivedItems, nextExpectedItem)

					nextExpectedItem++
				} else {
					break
				}
			}
		}
	}()

	for {
		w := <-workers
		if w == nil {
			deadcount++
			if deadcount == workercount {
				break
			}
			continue
		}
		pkg.ProcessWorkerQueue(w, CLIOptions, parameters, workers, workitem, dataitems)
		workitem++
	}

	if dataitems != nil {
		// Close the dataitems channel to signal that all work is dispatched
		close(dataitems)

		// Wait for all data items to be processed
		for len(receivedItems) > 0 {
			// Sleep for a short duration to avoid busy waiting
			time.Sleep(time.Millisecond)
		}
	}
}

func InitQueue(workflowCmd *cobra.Command) {
	queueCmd.Flags().BoolVarP(&CLIOptions.InlineImages, "inlineimages", "i", false, "Output images to terminal with Inline Image Protocol")
	queueCmd.Flags().BoolVarP(&CLIOptions.NoSaveData, "nosavedata", "n", false, "Do not save data to disk")
	queueCmd.Flags().StringVarP(&CLIOptions.OutputNodes, "outputnodes", "o", "", "Specify which output nodes save data. Comma separated nodes. Default is all nodes")

	// port to serve files on
	queueCmd.Flags().IntP("serveport", "", 8080, "File server port to serve files on")

	// storage path for uploaded files
	queueCmd.Flags().StringP("servepath", "", "", "Path to serve/reveive files from")

	// flag to indicate we should maintain order of the queue results
	queueCmd.Flags().BoolP("ordered", "", false, "Maintain the order of the queue results")

	workflowCmd.AddCommand(queueCmd)
}
