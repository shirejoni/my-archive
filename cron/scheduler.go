package cron

import (
	"errors"
	"fmt"
	"time"
)

type Scheduler interface {
	Next(t time.Time) time.Time
	Done() bool
}

type PeriodScheduler struct {
	period time.Duration
}

func (p *PeriodScheduler) Next(t time.Time) time.Time {
	return t.Truncate(time.Second).Add(p.period)
}

func (p *PeriodScheduler) Done() bool {
	return false
}

func (a *PeriodScheduler) At(t string) Scheduler {
	if a.period < time.Hour*24 {
		panic("Period Must be at least 1 Day")
	}
	h, m, err := parse(t)
	if err != nil {
		panic(err.Error())
	}
	return &AtScheduler{
		period: a.period,
		tH:     h,
		tM:     m,
	}
}

type AtScheduler struct {
	period time.Duration
	tH     int
	tM     int
}

func (a *AtScheduler) Next(t time.Time) time.Time {
	next := time.Date(t.Year(), t.Month(), t.Day(), a.tH, a.tM, 0, 0, time.UTC)
	if t.After(next) {
		return next.Add(a.period)
	}
	return next
}
func (a *AtScheduler) Done() bool {
	return false
}

type AtOnceScheduler struct {
	tH   int
	tM   int
	done bool
}

func (a *AtOnceScheduler) Next(t time.Time) time.Time {
	next := time.Date(t.Year(), t.Month(), t.Day(), a.tH, a.tM, 0, 0, t.Location())
	fmt.Println(next)
	if t.After(next) && a.done == false {
		next = next.Add(time.Hour * 24)
	}

	fmt.Println("again", next)
	a.done = true
	return next
}
func (a *AtOnceScheduler) Done() bool {
	return a.done
}

func AtOnce(t string) Scheduler {
	h, m, err := parse(t)
	fmt.Println("House", h, m)
	if err != nil {
		panic(err.Error())
	}
	return &AtOnceScheduler{
		tH:   h,
		tM:   m,
		done: false,
	}
}

func Every(t time.Duration) Scheduler {
	if t.Nanoseconds() < time.Second.Nanoseconds() {
		t = time.Second
	}
	t = t - time.Duration(t.Nanoseconds())%time.Second
	return &PeriodScheduler{
		period: t,
	}
}

func parse(t string) (int, int, error) {
	h := int(t[0]-'0')*10 + int(t[1]-'0')
	m := int(t[3]-'0')*10 + int(t[4]-'0')
	var err error
	if h < 0 || h > 24 {
		h, m = 0, 0
		err = errors.New("invalid h format")
	}
	if m < 0 || m > 59 {
		h, m = 0, 0
		err = errors.New("invalid m format")
	}
	return h, m, err
}
