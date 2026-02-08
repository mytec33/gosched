package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"git.sr.ht/~mytec/gosched/internal/logging"
	"git.sr.ht/~mytec/gosched/internal/schedule"
)

const (
	ExitSuccess      int = 0
	ExitNoConfig     int = 1
	ExitDecodeConfig int = 2
	ExitValidation   int = 3
	ExitInvalidArgs  int = 4
	ExitInvalidLock  int = 5
)

const defaultLockPort = 41037

var RunningWorkflows = schedule.NewSafeMapMutex()

type TickFunc func(now time.Time, data schedule.Schedule) int

func main() {
	os.Exit(run())
}

func run() int {
	logging.StdoutLogger.Info("startup", "event", "Scheduler service started")

	release, err := acquireSingleInstanceLock(defaultLockPort)
	if err != nil {
		logging.StderrLogger.Error("startup", "reason", "single-instance lock failed", "error", err)
		os.Exit(ExitInvalidLock)
	}
	defer release()
	logging.StdoutLogger.Info("startup", "event", "Acquired single-instance lock", "port", defaultLockPort)

	filename := ""
	summarizeConfig := false
	flag.StringVar(&filename, "schedule", "", "file containing a schedule to run")
	flag.BoolVar(&summarizeConfig, "summarize-config", false, "show concise summary of configuration file schedule")
	flag.Parse()

	if filename == "" {
		logging.StderrLogger.Error("invalid args")
		flag.Usage()
		os.Exit(ExitInvalidArgs)
	}

	schedules, err := schedule.ReadScheduleFile(filename)
	if err != nil {
		logging.StderrLogger.Error("startup", "reason", "failed to load schedule", "error", err)
		return ExitNoConfig
	}

	if summarizeConfig {
		displayConfigSummarization()
	}

	if errs := schedule.ValidateSchedule(schedules); len(errs) > 0 {
		for _, e := range errs {
			logging.StderrLogger.Error("startup", "configuration error", e)
		}
		return ExitValidation
	}

	logging.StdoutLogger.Info("startup", "Workflows loaded", len(schedules))
	fmt.Printf("Map with %%v: %v\n", schedules)

	runScheduler(runSchedulerTick, schedules)

	return 0
}

func displayConfigSummarization() {
	fmt.Println("display configuration summarization")

	os.Exit(ExitSuccess)
}

func acquireSingleInstanceLock(port int) (release func() error, err error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("another instance is likely running (cannot bind lock %s): %w", addr, err)
	}
	return ln.Close, nil
}

func runScheduler(tick TickFunc, s schedule.Schedule) {
	// Align to the next minute boundary once, then tick.
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().Truncate(time.Minute)
		n := tick(now, s)
		if n > 0 {
			logging.StdoutLogger.Info("scheduler", "scheduled workflows",
				n, "minute", now.Format("15:04"),
			)
		}
	}
}

func runSchedulerTick(now time.Time, wfm schedule.Schedule) int {
	currentMinute := schedule.MinuteKey(now.Format("15:04"))
	logging.StdoutLogger.Info("run scheduler tick", "current_minute", currentMinute)

	tasks := wfm[currentMinute]
	if len(tasks) == 0 {
		return 0
	}

	for _, task := range tasks {
		go executeWorkflow(task)
	}
	return len(tasks)
}

func executeWorkflow(wf schedule.Workflow) {
	workflowStart := time.Now()

	lockKey := wf.Name
	_, running := RunningWorkflows.Get(lockKey)
	if running {
		logging.StderrLogger.Error("execute", "workflow already running", lockKey)
		return
	}
	RunningWorkflows.Set(lockKey, "running")
	defer RunningWorkflows.Delete(lockKey)

	for i, step := range wf.Steps {
		stepStart := time.Now()

		args := strings.Fields(step.Args)
		logging.StdoutLogger.Info("execute", wf.Name, "starting", "stepIndex", i, "stepName", step.Name, "args", step.Args)

		cmd := exec.Command(step.Program, args...)
		output, err := cmd.CombinedOutput()
		stepDuration := time.Since(stepStart)

		if err != nil {
			logging.StderrLogger.Error("execute", wf.Name, "failed", "stepIndex", i, "stepName", step.Name, "args", step.Args, "duration", stepDuration, "reason", err, "output", string(output))
			return
		} else {
			logging.StdoutLogger.Info("execute", wf.Name, "completed", "stepIndex", i, "stepName", step.Name, "args", step.Args, "duration", stepDuration)
		}
	}

	workflowDuration := time.Since(workflowStart)
	logging.StdoutLogger.Info("execute", wf.Name, "completed", "duration", workflowDuration)
}
