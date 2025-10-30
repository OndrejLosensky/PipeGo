package main

import (
	"fmt"
	"log"
	"os"

	"pipego/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "run":
		configPath := "pipego.yml"
		if len(os.Args) >= 3 {
			configPath = os.Args[2]
		}
		if err := cmd.Run(configPath); err != nil {
			log.Fatal(err)
		}
	case "serve":
		if err := cmd.Serve(); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: pipego [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  run [config-path]    Run a pipeline")
	fmt.Println("  serve                Start HTTP server")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  pipego run ../dummy-app/pipego.yml")
	fmt.Println("  pipego serve")
}