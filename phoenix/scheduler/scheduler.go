package scheduler

import (
	"challenge-to-you/phoenix/config"
	"challenge-to-you/phoenix/pipeline"
	"log"
	"time"
)

type Job struct {
	Name     string
	Interval time.Duration
	Action   func() error
}

type PipelineScheduler struct {
	Config   *config.Config
	Pipeline *pipeline.RepairPipeline
	Jobs     []*Job
	StopChan chan struct{}
}

func NewPipelineScheduler(cfg *config.Config, pipe *pipeline.RepairPipeline) *PipelineScheduler {
	return &PipelineScheduler{
		Config:   cfg,
		Pipeline: pipe,
		Jobs:     make([]*Job, 0),
		StopChan: make(chan struct{}),
	}
}

func (s *PipelineScheduler) RegisterJob(name string, interval time.Duration, action func() error) {
	s.Jobs = append(s.Jobs, &Job{
		Name:     name,
		Interval: interval,
		Action:   action,
	})
}

func (s *PipelineScheduler) Start() {
	log.Printf("Starting Project Phoenix Pipeline Scheduler daemon...")
	for _, job := range s.Jobs {
		go func(j *Job) {
			ticker := time.NewTicker(j.Interval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					log.Printf("Executing scheduled task: %s", j.Name)
					if err := j.Action(); err != nil {
						log.Printf("Error running scheduled task %s: %v", j.Name, err)
					}
				case <-s.StopChan:
					return
				}
			}
		}(job)
	}
}

func (s *PipelineScheduler) Stop() {
	log.Printf("Stopping Pipeline Scheduler...")
	close(s.StopChan)
}
