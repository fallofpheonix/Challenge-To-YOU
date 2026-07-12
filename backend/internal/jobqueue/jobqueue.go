package jobqueue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type JobType int

const (
	JobAIRequest JobType = iota
	JobSandboxExec
	JobDatabaseOp
)

func (jt JobType) String() string {
	switch jt {
	case JobAIRequest:
		return "ai_request"
	case JobSandboxExec:
		return "sandbox_exec"
	case JobDatabaseOp:
		return "database_op"
	default:
		return "unknown"
	}
}

type JobResult struct {
	Type      JobType
	Payload   interface{}
	Err       error
	Duration  time.Duration
	Completed bool
}

type Job struct {
	Type      JobType
	Priority  int
	Execute   func(context.Context) (interface{}, error)
	CreatedAt time.Time
	ID        int64
}

type Queue struct {
	mu       sync.Mutex
	jobs     []*Job
	nextID   int64
	workers  int
	resultCh chan JobResult
	closed   bool
	wg       sync.WaitGroup
}

func New(workers int) *Queue {
	if workers <= 0 {
		workers = 4
	}
	q := &Queue{
		workers:  workers,
		resultCh: make(chan JobResult, 100),
	}

	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	return q
}

func (q *Queue) worker(id int) {
	defer q.wg.Done()
	for {
		q.mu.Lock()
		if q.closed && len(q.jobs) == 0 {
			q.mu.Unlock()
			return
		}
		if len(q.jobs) == 0 {
			q.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}

		bestIdx := 0
		for i, job := range q.jobs {
			if job.Priority > q.jobs[bestIdx].Priority {
				bestIdx = i
			}
		}
		job := q.jobs[bestIdx]
		q.jobs = append(q.jobs[:bestIdx], q.jobs[bestIdx+1:]...)
		q.mu.Unlock()

		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		result, err := job.Execute(ctx)
		cancel()

		duration := time.Since(start)
		log.Printf("[JOBQUEUE] worker=%d job=%s id=%d duration=%v err=%v", id, job.Type, job.ID, duration, err)

		select {
		case q.resultCh <- JobResult{
			Type:      job.Type,
			Payload:   result,
			Err:       err,
			Duration:  duration,
			Completed: err == nil,
		}:
		default:
		}
	}
}

func (q *Queue) Enqueue(jobType JobType, priority int, fn func(context.Context) (interface{}, error)) (int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return 0, fmt.Errorf("job queue is closed")
	}

	q.nextID++
	job := &Job{
		Type:      jobType,
		Priority:  priority,
		Execute:   fn,
		CreatedAt: time.Now(),
		ID:        q.nextID,
	}

	q.jobs = append(q.jobs, job)
	return job.ID, nil
}

func (q *Queue) Results() <-chan JobResult {
	return q.resultCh
}

func (q *Queue) Pending() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.jobs)
}

func (q *Queue) Shutdown() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	q.wg.Wait()
}

type Client struct {
	queue *Queue
}

func NewClient(queue *Queue) *Client {
	return &Client{queue: queue}
}

func (c *Client) RunAI(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	id, err := c.queue.Enqueue(JobAIRequest, 1, fn)
	if err != nil {
		return nil, err
	}
	return c.waitForResult(ctx, id)
}

func (c *Client) RunSandbox(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	id, err := c.queue.Enqueue(JobSandboxExec, 2, fn)
	if err != nil {
		return nil, err
	}
	return c.waitForResult(ctx, id)
}

func (c *Client) RunDB(ctx context.Context, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	id, err := c.queue.Enqueue(JobDatabaseOp, 0, fn)
	if err != nil {
		return nil, err
	}
	return c.waitForResult(ctx, id)
}

func (c *Client) waitForResult(ctx context.Context, jobID int64) (interface{}, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-c.queue.Results():
			if result.Err != nil {
				return nil, result.Err
			}
			return result.Payload, nil
		}
	}
}
