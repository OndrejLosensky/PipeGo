package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Project represents a project configuration
type Project struct {
	Name        string `yaml:"name" json:"name"`
	Path        string `yaml:"path" json:"path"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// ProjectsConfig holds the list of all projects
type ProjectsConfig struct {
	Projects []Project `yaml:"projects" json:"projects"`
}

// LoadProjects loads the projects configuration from a YAML file
func LoadProjects(configPath string) (*ProjectsConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects config: %w", err)
	}

	var config ProjectsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse projects config: %w", err)
	}

	return &config, nil
}

// GetProject returns a project by name
func (pc *ProjectsConfig) GetProject(name string) (*Project, error) {
	for _, project := range pc.Projects {
		if project.Name == name {
			return &project, nil
		}
	}
	return nil, fmt.Errorf("project '%s' not found", name)
}

// ValidateProject checks if a project's path exists and has a pipego.yml
func (p *Project) Validate(baseDir string) error {
	// Make path absolute if relative
	projectPath := p.Path
	if !filepath.IsAbs(projectPath) {
		projectPath = filepath.Join(baseDir, projectPath)
	}

	// Check if directory exists
	info, err := os.Stat(projectPath)
	if err != nil {
		return fmt.Errorf("project path does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("project path is not a directory")
	}

	// Check if pipego.yml exists
	pipegoPath := filepath.Join(projectPath, "pipego.yml")
	if _, err := os.Stat(pipegoPath); err != nil {
		return fmt.Errorf("pipego.yml not found in project directory")
	}

	return nil
}

// GetPipegoPath returns the absolute path to the project's pipego.yml
func (p *Project) GetPipegoPath(baseDir string) string {
	projectPath := p.Path
	if !filepath.IsAbs(projectPath) {
		projectPath = filepath.Join(baseDir, projectPath)
	}
	return filepath.Join(projectPath, "pipego.yml")
}

