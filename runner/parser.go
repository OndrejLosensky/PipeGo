package runner

import (
    "os"
    "gopkg.in/yaml.v3"
)

type Step struct {
    Name string `yaml:"name"`
    Run  string `yaml:"run"`
}

type Config struct {
    Steps []Step `yaml:"steps"`
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
