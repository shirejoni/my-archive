package main

import (
	"fmt"
	"github.com/shirejoni/my-archive/cron"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func cronStart(c *cron.Cron, cronStopped chan bool) {
	c.Add(cron.AtOnce("15:13"), func() {
		fmt.Println("Hello From Second", time.Now().Second())
	})
	c.Start()
	c.Wait()
	cronStopped <- true
}

func combineStopAndSignals(stop chan bool, signals chan os.Signal) chan bool {
	out := make(chan bool)
	go func() {
		select {
		case <-stop:
			out <- true
		case <-signals:
			out <- true
		}
	}()
	return out
}

func main() {
	// Channels that terminate crons
	stop := make(chan bool)
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	c := cron.New()
	go cronStart(c, stop)

mainLoop:
	for {
		select {
		case <-combineStopAndSignals(stop, signals):
			fmt.Println("Cron is Finished")
			c.Stop()
			break mainLoop
		}
	}

}
