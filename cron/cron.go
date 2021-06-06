package cron

import (
	"sort"
	"time"
)

type JobFunc func()

func (j JobFunc) Run() {
	j()
}

type Job struct {
	Scheduler Scheduler
	Callback  JobFunc
	Next      time.Time
	Prev      time.Time
	Active    bool
}
type byTime []*Job

func (b byTime) Len() int {
	return len(b)
}
func (b byTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byTime) Less(i, j int) bool {
	if b[i].Next.IsZero() {
		return false
	}
	if b[j].Next.IsZero() {
		return true
	}

	return b[i].Next.Before(b[j].Next)
}

type Cron struct {
	jobs    []*Job
	running bool
	add     chan *Job
	stop    chan struct{}
}

func New() *Cron {
	return &Cron{
		add:  make(chan *Job),
		stop: make(chan struct{}),
	}
}

func (c *Cron) Start() {
	c.running = true
	go c.run()
}

func (c *Cron) Stop() {
	if !c.running {
		return
	}
	c.running = false
	c.stop <- struct{}{}
}

func (c *Cron) Wait() {
	<-c.stop
}

func (c *Cron) Add(s Scheduler, j func()) {
	job := &Job{
		Scheduler: s,
		Callback:  JobFunc(j),
		Active:    true,
	}
	if !c.running {
		c.jobs = append(c.jobs, job)
		return
	}
	c.add <- job
}

func (c *Cron) run() {
	var nextEmit time.Time
	now := time.Now()
	for _, j := range c.jobs {
		j.Next = j.Scheduler.Next(now)
	}
	for {
		now := time.Now()
		sort.Sort(byTime(c.jobs))
		if len(c.jobs) > 0 {
			nextEmit = c.jobs[0].Next
		} else {
			nextEmit = now.AddDate(10, 0, 0)
		}

		select {
		case now = <-time.After(nextEmit.Sub(now)):
			for index, job := range c.jobs {
				if job.Next != nextEmit {
					break
				}
				job.Prev = now
				job.Next = job.Scheduler.Next(now)
				go job.Callback.Run()
				if job.Scheduler.Done() {
					c.jobs = append(c.jobs[:index], c.jobs[index+1:]...)
					if len(c.jobs) == 0 {
						return
					}
					continue
				}

			}

		case j := <-c.add:
			j.Next = j.Scheduler.Next(time.Now())
			c.jobs = append(c.jobs, j)
		case <-c.stop:
			return
		}
	}

}
