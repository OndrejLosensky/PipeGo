package main

import (
	"fmt"
	"log"
	"os"

	"pipego/cmd"
)

func main() {
	if len(os.Args) < 2 {
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
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}