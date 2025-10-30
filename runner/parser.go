package runner

import (
    "fmt"
    "os"
    "strings"
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

type Group struct {
    Parts map[string]Part `yaml:"parts"`
}

type Schedule struct {
    Parts  []string `yaml:"parts,omitempty"`  // "frontend.deploy" or "old-part"
    Groups []string `yaml:"groups,omitempty"` // "frontend" runs all parts in group
    At     string   `yaml:"at,omitempty"`
    Every  string   `yaml:"every,omitempty"`
}

type Config struct {
    // Backward compatibility: support old format with direct steps array
    Steps []Step `yaml:"steps,omitempty"`
    // Old format: support flat parts map
    Parts map[string]Part `yaml:"parts,omitempty"`
    // New format: support hierarchical groups
    Groups map[string]Group `yaml:"groups,omitempty"`
    // Schedules for automatic runs
    Schedules []Schedule `yaml:"schedules,omitempty"`
}

// GetAllParts returns all parts with their steps
// Flattens groups to "group.part" format (e.g., "frontend.deploy")
// For backward compatibility, if no parts/groups are defined, returns a single "default" part
func (c *Config) GetAllParts() map[string][]Step {
    result := make(map[string][]Step)
    
    // Add grouped parts with "group.part" naming
    for groupName, group := range c.Groups {
        for partName, part := range group.Parts {
            fullPath := fmt.Sprintf("%s.%s", groupName, partName)
            result[fullPath] = part.Steps
        }
    }
    
    // Add flat parts (old format or ungrouped parts)
    for partName, part := range c.Parts {
        result[partName] = part.Steps
    }
    
    // Oldest format: wrap steps in a "default" part
    if len(result) == 0 && len(c.Steps) > 0 {
        result["default"] = c.Steps
    }
    
    return result
}

// GetGroup returns all parts within a specific group
func (c *Config) GetGroup(groupName string) (map[string][]Step, error) {
    group, exists := c.Groups[groupName]
    if !exists {
        return nil, fmt.Errorf("group '%s' not found", groupName)
    }
    
    result := make(map[string][]Step)
    for partName, part := range group.Parts {
        fullPath := fmt.Sprintf("%s.%s", groupName, partName)
        result[fullPath] = part.Steps
    }
    
    return result, nil
}

// GetPart returns steps for a specific part
// Supports both flat names ("tests") and grouped names ("frontend.deploy")
func (c *Config) GetPart(partName string) ([]Step, error) {
    parts := c.GetAllParts()
    steps, exists := parts[partName]
    if !exists {
        return nil, fmt.Errorf("part '%s' not found", partName)
    }
    return steps, nil
}

// ParsePartName splits a part name into group and part components
// Returns ("", partName) for ungrouped parts, (groupName, partName) for grouped parts
func ParsePartName(fullPath string) (string, string) {
    parts := strings.SplitN(fullPath, ".", 2)
    if len(parts) == 2 {
        return parts[0], parts[1]
    }
    return "", fullPath
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
