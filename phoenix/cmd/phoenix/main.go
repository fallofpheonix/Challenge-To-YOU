package main

import (
	"challenge-to-you/phoenix/config"
	"challenge-to-you/phoenix/pipeline"
	"challenge-to-you/phoenix/scheduler"
	"flag"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	configFlag := flag.String("config", "phoenix_config.json", "path to configuration file")
	runOnceFlag := flag.Bool("run-once", false, "run validation and exit")
	flag.Parse()

	log.Println("Initializing Project Phoenix Autonomous Software Engineering Daemon...")

	cfg, err := config.LoadConfig(*configFlag)
	if err != nil {
		log.Fatalf("Fatal: failed to load configuration: %v", err)
	}

	pipe := pipeline.NewRepairPipeline(cfg)

	if *runOnceFlag {
		log.Println("Executing single validation scan...")
		if errVal := pipe.Validators.VerifyAll(); errVal != nil {
			log.Printf("Scan completed with errors:\n%v", errVal)
			// Trigger single mock repair on compiler failure test files
			_ = pipe.RunSelfRepair(errVal.Error(), "backend/main.go")
		} else {
			log.Println("Scan completed: repository is clean.")
		}
		return
	}

	// Initialize Scheduler
	sched := scheduler.NewPipelineScheduler(cfg, pipe)

	// Register File Watcher Check
	var lastModified time.Time
	sched.RegisterJob("File Change Detector", time.Duration(cfg.WatchIntervalMs)*time.Millisecond, func() error {
		modified := false
		errWalk := filepath.WalkDir(cfg.WorkspaceRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				// Skip git, doc folders, and build outputs
				if d.Name() == ".git" || d.Name() == "docs" || d.Name() == "brain" || d.Name() == "phoenix" {
					return filepath.SkipDir
				}
				return nil
			}

			// Audit file time
			info, errStat := d.Info()
			if errStat != nil {
				return nil
			}
			if info.ModTime().After(lastModified) {
				if lastModified.Year() > 2000 {
					modified = true
					log.Printf("Detected modifications in source file: %s", path)
				}
				lastModified = info.ModTime()
			}
			return nil
		})

		if errWalk != nil {
			return errWalk
		}

		if modified {
			log.Println("File change detected. Executing repository verification check...")
			if errVal := pipe.Validators.VerifyAll(); errVal != nil {
				log.Printf("Repository errors found: %v", errVal)
				// Extract primary error log details
				errMsg := errVal.Error()
				targetFile := extractFailingFile(errMsg)
				_ = pipe.RunSelfRepair(errMsg, targetFile)
			} else {
				log.Println("Verification passed. Codebase is clean.")
			}
		}

		return nil
	})

	// Start Daemon
	sched.Start()

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	sched.Stop()
	log.Println("Project Phoenix Daemon exited clean.")
}

func extractFailingFile(errMsg string) string {
	// Simple path extraction helper
	idx := strings.Index(errMsg, ".go:")
	if idx == -1 {
		return "backend/main.go"
	}
	start := strings.LastIndex(errMsg[:idx], " ")
	if start == -1 {
		start = 0
	}
	return strings.TrimSpace(errMsg[start : idx+3])
}
