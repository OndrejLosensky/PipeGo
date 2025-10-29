package main

import (
    "fmt"
    "log"
    "os"
    "pipego/runner"
)

func main() {
    command := os.Args[1]
    if command == "run" {
        configPath := "pipego.yml"
        if len(os.Args) >= 3 {
            configPath = os.Args[2]
        }
        err := runner.RunPipeline(configPath)
        if err != nil {
            log.Fatalf("Pipeline failed: %v", err)
        }
    } else {
        fmt.Println("Unknown command:", command)
        os.Exit(1)
    }
}
