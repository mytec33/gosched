package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Job struct {
	Time string   `json:"time"`
	Prog string   `json:"prog"`
	Args []string `json:"args"`
	Wdir string   `json:"wdir"`
}

func main() {
	log.Println("Scheduler service started")

	// Load schedule from JSON file
	schedule, err := loadSchedule("schedule.json")
	if err != nil {
		log.Fatalf("Failed to load schedule: %v", err)
	}

	log.Printf("Loaded %d jobs from schedule.json", len(schedule))

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
				log.Printf("Scheduled %d job(s) for %s", jobsThisMinute, currentMinute)
			}

			lastMinute = currentMinute
		}

		// Sleep until next minute boundary
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
	start := time.Now()
	args := strings.Join(job.Args, " ")
	log.Printf("Starting job: %s", args)

	cmd := exec.Command("your_program.exe", job.Args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	if err != nil {
		log.Printf("Job FAILED (%v): %s - Error: %v - Output: %s",
			duration, args, err, string(output))
	} else {
		log.Printf("Job completed (%v): %s", duration, args)
	}
}
