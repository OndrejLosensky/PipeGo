package runner

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

func RunPipeline(configPath string) error {
    cfg, err := LoadConfig(configPath)
    if err != nil {
        return err
    }

    // Change to the directory where runner.yml is located
    configDir := filepath.Dir(configPath)
    originalDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("failed to get current directory: %w", err)
    }
    
    err = os.Chdir(configDir)
    if err != nil {
        return fmt.Errorf("failed to change to config directory: %w", err)
    }
    defer os.Chdir(originalDir) 

    for _, step := range cfg.Steps {
        fmt.Println("→", step.Name)
        cmd := exec.Command("bash", "-c", step.Run)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        err := cmd.Run()
        if err != nil {
            fmt.Println("❌ Step failed:", err)
            return err
        }
        fmt.Println("✅ Done:", step.Name)
    }
    fmt.Println("🏁 All steps finished successfully.")
    return nil
}
