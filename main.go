package main

import (
	"log"
	"os"

	"github.com/creator54/temporal-go-demo-app/starter"
	"github.com/creator54/temporal-go-demo-app/workers"
)

func main() {
	// Check if we're starting a workflow or running a worker
	if len(os.Args) > 1 && os.Args[1] == "start" {
		// Get the name from command line args or use default
		name := "Temporal"
		if len(os.Args) > 2 {
			name = os.Args[2]
		}

		// Start the workflow
		err := starter.StartWorkflow(name)
		if err != nil {
			log.Fatalf("Failed to start workflow: %v", err)
		}
	} else {
		// Start the worker
		err := workers.StartWorker()
		if err != nil {
			log.Fatalf("Failed to start worker: %v", err)
		}
	}
} 