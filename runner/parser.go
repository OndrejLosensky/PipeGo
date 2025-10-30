package runner

import (
    "fmt"
    "os"
    "gopkg.in/yaml.v3"
)

type Step struct {
    Name     string `yaml:"name"`
    Run      string `yaml:"run"`
    Category string `yaml:"category,omitempty"` // Optional category (tests, deploy, setup, etc.)
}

type Part struct {
    Steps []Step `yaml:"steps"`
}

type Config struct {
    // Backward compatibility: support old format with direct steps array
    Steps []Step `yaml:"steps,omitempty"`
    // New format: support parts map
    Parts map[string]Part `yaml:"parts,omitempty"`
}

// GetAllParts returns all parts with their steps
// For backward compatibility, if no parts are defined, returns a single "default" part
func (c *Config) GetAllParts() map[string][]Step {
    if len(c.Parts) > 0 {
        // New format: return all parts
        result := make(map[string][]Step)
        for partName, part := range c.Parts {
            result[partName] = part.Steps
        }
        return result
    }
    
    // Old format: wrap steps in a "default" part
    if len(c.Steps) > 0 {
        return map[string][]Step{
            "default": c.Steps,
        }
    }
    
    return make(map[string][]Step)
}

// GetPart returns steps for a specific part
func (c *Config) GetPart(partName string) ([]Step, error) {
    parts := c.GetAllParts()
    steps, exists := parts[partName]
    if !exists {
        return nil, fmt.Errorf("part '%s' not found", partName)
    }
    return steps, nil
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg Config
    err = yaml.Unmarshal(data, &cfg)
    if err != nil {
        return nil, err
    }
    return &cfg, nil
}
