package scheduler

import (
	"fmt"
	"sync"
	"time"
)

type jobEntry struct {
	job      Job
	nextRun  time.Time
	cron     *CronSchedule
	interval time.Duration
}

type JobRegistry struct {
	mu     sync.RWMutex
	items  map[string]*jobEntry
	order  []string
}

func NewJobRegistry() *JobRegistry {
	return &JobRegistry{items: make(map[string]*jobEntry)}
}

func (r *JobRegistry) Register(job Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[job.Name]; ok {
		return fmt.Errorf("job already registered: %s", job.Name)
	}

	entry := &jobEntry{job: job}

	if dur, err := time.ParseDuration(job.Schedule); err == nil {
		entry.interval = dur
		entry.nextRun = time.Now().Add(dur)
	} else {
		cron, err := ParseCron(job.Schedule)
		if err != nil {
			return fmt.Errorf("invalid schedule %q for job %s: %w", job.Schedule, job.Name, err)
		}
		entry.cron = cron
		entry.nextRun = cron.Next(time.Now())
	}

	r.items[job.Name] = entry
	r.order = append(r.order, job.Name)
	return nil
}

func (r *JobRegistry) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[name]; !ok {
		return fmt.Errorf("job not found: %s", name)
	}

	delete(r.items, name)
	for i, n := range r.order {
		if n == name {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
	return nil
}

func (r *JobRegistry) Due(now time.Time) []*jobEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var due []*jobEntry
	for _, name := range r.order {
		entry := r.items[name]
		if entry.nextRun.Before(now) || entry.nextRun.Equal(now) {
			due = append(due, entry)
			if entry.interval > 0 {
				entry.nextRun = now.Add(entry.interval)
			} else if entry.cron != nil {
				entry.nextRun = entry.cron.Next(now)
			}
		}
	}
	return due
}

func (r *JobRegistry) All() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.order))
	copy(out, r.order)
	return out
}
