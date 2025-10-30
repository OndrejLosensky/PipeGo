package runner

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"pipego/events"
	"pipego/runner/storage"
)

// Scheduler manages automatic pipeline runs based on schedules
type Scheduler struct {
	projectsConfig *ProjectsConfig
	storage        *storage.Storage
	baseDir        string
	stopChan       chan struct{}
	lastRuns       map[string]time.Time // track last execution per schedule
	mu             sync.RWMutex         // protect lastRuns map
	runningJobs    map[string]bool      // track currently running schedules
}

// NewScheduler creates a new scheduler instance
func NewScheduler(projectsConfig *ProjectsConfig, storage *storage.Storage, baseDir string) *Scheduler {
	return &Scheduler{
		projectsConfig: projectsConfig,
		storage:        storage,
		baseDir:        baseDir,
		stopChan:       make(chan struct{}),
		lastRuns:       make(map[string]time.Time),
		runningJobs:    make(map[string]bool),
	}
}

// Start begins the scheduler loop
func (s *Scheduler) Start() {
	log.Println("📅 Scheduler started")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Run tick immediately on start
	s.tick()

	for {
		select {
		case <-ticker.C:
			s.tick()
		case <-s.stopChan:
			log.Println("📅 Scheduler stopped")
			return
		}
	}
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
}

// tick checks all schedules and triggers runs if needed
func (s *Scheduler) tick() {
	for _, project := range s.projectsConfig.Projects {
		configPath := project.GetPipegoPath(s.baseDir)
		
		// Load config
		cfg, err := LoadConfig(configPath)
		if err != nil {
			// Skip if config can't be loaded (might not have schedules)
			continue
		}

		// No schedules defined
		if len(cfg.Schedules) == 0 {
			continue
		}

		// Check each schedule
		for i, schedule := range cfg.Schedules {
			scheduleKey := fmt.Sprintf("%s-schedule-%d", project.Name, i)
			
			// Check if schedule should run
			s.mu.RLock()
			lastRun := s.lastRuns[scheduleKey]
			isRunning := s.runningJobs[scheduleKey]
			s.mu.RUnlock()

			// Skip if already running
			if isRunning {
				continue
			}

			if s.shouldRun(schedule, lastRun) {
				// Validate parts exist
				if len(schedule.Parts) > 0 {
					allParts := cfg.GetAllParts()
					for _, partName := range schedule.Parts {
						if _, exists := allParts[partName]; !exists {
							log.Printf("⚠️  Schedule skipped: part '%s' not found in %s", partName, project.Name)
							continue
						}
					}
				}

				// Mark as running
				s.mu.Lock()
				s.runningJobs[scheduleKey] = true
				s.lastRuns[scheduleKey] = time.Now()
				s.mu.Unlock()

				// Execute in goroutine
				go func(p Project, sched Schedule, key string) {
					s.executeSchedule(p.Name, sched)
					
					// Mark as not running
					s.mu.Lock()
					delete(s.runningJobs, key)
					s.mu.Unlock()
				}(project, schedule, scheduleKey)
			}
		}
	}
}

// shouldRun determines if a schedule should be triggered now
func (s *Scheduler) shouldRun(schedule Schedule, lastRun time.Time) bool {
	now := time.Now()

	// Time-based schedule (at: "HH:MM")
	if schedule.At != "" {
		hour, minute, err := parseAtTime(schedule.At)
		if err != nil {
			log.Printf("⚠️  Invalid time format '%s': %v", schedule.At, err)
			return false
		}

		// Check if current time matches schedule time
		if now.Hour() == hour && now.Minute() == minute {
			// Ensure we only run once per day at this time
			if lastRun.IsZero() || now.Sub(lastRun) >= 23*time.Hour {
				return true
			}
		}
		return false
	}

	// Interval-based schedule (every: "1h", "30m", etc.)
	if schedule.Every != "" {
		interval, err := parseInterval(schedule.Every)
		if err != nil {
			log.Printf("⚠️  Invalid interval format '%s': %v", schedule.Every, err)
			return false
		}

		// First run or interval elapsed
		if lastRun.IsZero() || now.Sub(lastRun) >= interval {
			return true
		}
		return false
	}

	return false
}

// executeSchedule triggers a pipeline run for the given schedule
func (s *Scheduler) executeSchedule(projectName string, schedule Schedule) {
	project, err := s.projectsConfig.GetProject(projectName)
	if err != nil {
		log.Printf("❌ Schedule execution failed: %v", err)
		return
	}

	configPath := project.GetPipegoPath(s.baseDir)
	partsStr := "all parts"
	if len(schedule.Parts) > 0 {
		partsStr = strings.Join(schedule.Parts, ", ")
	}

	scheduleType := schedule.At
	if scheduleType == "" {
		scheduleType = schedule.Every
	}

	log.Printf("⏰ Schedule triggered: %s (parts: %s) - %s", projectName, partsStr, scheduleType)

	// Broadcast event to SSE clients
	broker := events.GetBroker()
	broker.Broadcast("run_started", map[string]interface{}{
		"project": projectName,
		"parts":   schedule.Parts,
		"type":    "scheduled",
	})

	// If no parts specified, run all parts
	if len(schedule.Parts) == 0 {
		_, err := RunPipelineWithOptions(configPath, RunPipelineOptions{
			Storage:          s.storage,
			StreamToTerminal: false,
			PartFilter:       "",
		})
		if err != nil {
			log.Printf("❌ Scheduled run failed for %s: %v", projectName, err)
		} else {
			log.Printf("✅ Scheduled run completed: %s", projectName)
		}
		return
	}

	// Run specific parts
	for _, partName := range schedule.Parts {
		_, err := RunPipelineWithOptions(configPath, RunPipelineOptions{
			Storage:          s.storage,
			StreamToTerminal: false,
			PartFilter:       partName,
		})
		if err != nil {
			log.Printf("❌ Scheduled run failed for %s (%s): %v", projectName, partName, err)
		} else {
			log.Printf("✅ Scheduled run completed: %s (%s)", projectName, partName)
		}
	}
}

// parseAtTime parses "HH:MM" format
func parseAtTime(at string) (hour, minute int, err error) {
	parts := strings.Split(at, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format, expected HH:MM")
	}

	var hourVal int
	hourVal, err = strconv.Atoi(parts[0])
	if err != nil || hourVal < 0 || hourVal > 23 {
		return 0, 0, fmt.Errorf("invalid hour")
	}
	hour = hourVal

	minute, err = strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid minute")
	}

	return hour, minute, nil
}

// parseInterval parses duration strings like "1h", "30m", "1h30m"
func parseInterval(every string) (time.Duration, error) {
	// Handle combined formats like "1h30m"
	if strings.Contains(every, "h") && strings.Contains(every, "m") {
		re := regexp.MustCompile(`(\d+)h(\d+)m`)
		matches := re.FindStringSubmatch(every)
		if len(matches) == 3 {
			hours, _ := strconv.Atoi(matches[1])
			minutes, _ := strconv.Atoi(matches[2])
			return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute, nil
		}
	}

	// Simple formats like "1h", "30m", "24h"
	duration, err := time.ParseDuration(every)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format")
	}

	return duration, nil
}

