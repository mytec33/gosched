package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"git.sr.ht/~mytec/gosched/pkg/logging"
)

const (
	ExitNoConfig int = 1
)

type Job struct {
	Name  string `json:"name"`
	Time  string `json:"time"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Name    string `json:"name"`
	Program string `json:"program"`
	Args    string `json:"args"`
}

var PropertyPaths = map[string]string{
	"BAC": "e:\\Vitruvian\\retail-to-vitruvian-export.exe",
	"CHI": "e:\\Vitruvian\\retail-to-vitruvian-export.exe",
	"DDI": "d:\\vitruvian\\retail-to-vitruvian-export.exe",
	"EVN": "e:\\Vitruvian-Evansville\\retail-to-vitruvian-export.exe",
	"SHR": "q:\\Vitruvian\\retail-to-vitruvian-export.exe",
	"TAH": "e:\\Vitruvian-Tahoe\\retail-to-vitruvian-export.exe",
	"TRL": "e:\\Vitruvian\\retail-to-vitruvian-export.exe",
	"TRT": "e:\\Vitruvian\\retail-to-vitruvian-export.exe",
}

var RunningJobs = &SafeMapRWMutex{
	data: make(map[string]string),
}

type SafeMapRWMutex struct {
	mu   sync.RWMutex
	data map[string]string
}

func (sm *SafeMapRWMutex) Delete(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.data, key)
}

func (sm *SafeMapRWMutex) Set(key string, value string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data[key] = value
}

func (sm *SafeMapRWMutex) Get(key string) (string, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	val, ok := sm.data[key]

	return val, ok
}

func main() {
	logging.StdoutLogger.Info("startup", "Scheduler service started", "")

	schedule, err := loadSchedule("schedule.json")
	if err != nil {
		logging.StderrLogger.Error("startup", "reason", "failed to load schedule", "error", err)
		os.Exit(ExitNoConfig)
	}

	logging.StderrLogger.Error("startup", "Jobs loaded", len(schedule))

	lastMinute := ""

	for {
		now := time.Now()
		currentMinute := now.Format("15:04")

		if currentMinute != lastMinute {
			jobsThisMinute := 0
			for _, job := range schedule {
				if job.Time == currentMinute {
					jobsThisMinute++
					go executeJob(job)
				}
			}

			if jobsThisMinute > 0 {
				logging.StderrLogger.Error("scheduler", "scheduled job(s)", jobsThisMinute, "minute", currentMinute)
			}

			lastMinute = currentMinute
		}

		nextMinute := now.Truncate(time.Minute).Add(time.Minute)
		time.Sleep(time.Until(nextMinute))
	}
}

func loadSchedule(filename string) ([]Job, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var schedule []Job
	err = json.Unmarshal(data, &schedule)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return schedule, nil
}

func executeJob(job Job) {
	telemetryStart := time.Now()

	jobLock := job.Name
	_, running := RunningJobs.Get(jobLock)
	if running {
		logging.StderrLogger.Error("execute", "job already running", jobLock)
		return
	}
	RunningJobs.Set(jobLock, "running")
	defer RunningJobs.Delete(jobLock)

	for i, step := range job.Steps {
		args := strings.Fields(step.Args)
		if len(args) == 0 {
			logging.StderrLogger.Error("execute", "job", job.Name, "step", i, "has no args", step.Name)
			return
		}

		if len(args) < 2 {
			logging.StderrLogger.Error("execute", "job", job.Name, "step", i, "too few args", step.Name)
			return
		}
		propertyCode := args[1] // This needs to be made more generic

		program, ok := PropertyPaths[propertyCode]
		if !ok {
			logging.StderrLogger.Error("execute", "job no .exe found for property", propertyCode)
			return
		}

		logging.StdoutLogger.Info("execute", job.Name, "starting", "step", i, "args", step.Args)

		cmd := exec.Command(program, args...)
		output, err := cmd.CombinedOutput()
		duration := time.Since(telemetryStart)

		if err != nil {
			logging.StderrLogger.Error("execute", job.Name, "failed", "step", i, "args", step.Args, "duration", duration, "reason", err, "output", string(output))
			break
		} else {
			logging.StdoutLogger.Info("execute", job.Name, "completed", "step", i, "args", step.Args, "duration", duration)
		}
	}
}
