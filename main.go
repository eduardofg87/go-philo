package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/pkg/namesgenerator"
	"log"
	"math/rand"
	"os"
	"time"
)

var (
	THINK_MAX_TIME  time.Duration
	EAT_TIME        time.Duration
	HUNGRY_MAX_TIME time.Duration
	PHILOS          int
	duration        time.Duration
	names           = []string{}
)

type fork bool // Used/Free For future improvement

type announcement struct {
	from, message string
}

func (a announcement) String() string {
	return fmt.Sprintf("%-25s: %s", a.from, a.message)
}

type philosopher struct {
	name     string
	left     *fork
	right    *fork
	state    string
	dying    chan int
	announce chan announcement
}

func (p *philosopher) timeTrack(from time.Time, method string) {
	p.say(fmt.Sprintf("I've finished %s in %vs. I'm now %s.", method, int64(time.Since(from))/1e9, p.state))
}

func (p *philosopher) say(message string) {
	p.announce <- announcement{from: p.name, message: message}
}

func (p philosopher) Live() {

	defer p.timeTrack(time.Now(), "my life")
	defer func() {
		p.state = "dead"
		p.dying <- 1
	}()
	p.state = "hungry"
	for p.state != "dead" {
		from := time.Now()
		func() {
			defer p.timeTrack(from, p.state)
			switch p.state {
			case "think":
				time.Sleep(time.Duration(rand.Intn(int(THINK_MAX_TIME)) + 2))
				p.state = "hungry"
			case "dead":
				return
			case "hungry":
				for time.Since(from) < HUNGRY_MAX_TIME {
					if *p.left && *p.right {
						*p.left, *p.right = false, false
						p.state = "eat"
						return
					}
					time.Sleep(200 * time.Millisecond)
				}
				p.state = "dead"
			case "eat":
				time.Sleep(EAT_TIME)
				*p.left = true
				*p.right = true
				p.state = "think"
			}
		}()
	}
}

//Random Stuff

func timeTrack(from time.Time, name string, a chan announcement) {
	a <- announcement{from: "Main", message: fmt.Sprintf("Execution of %s took %v", name, time.Since(from))}
}

func watcher(c chan announcement) {
	go func() {
		for a := range c {
			log.Println(a)
		}
	}()
}

func Run(c *cli.Context) {
	a := make(chan announcement)
	defer timeTrack(time.Now(), "main", a)
	d := make(chan int)
	phils := []philosopher{}
	forks := []fork{}
	for i := PHILOS; i >= 0; i-- {
		names = append(names, namesgenerator.GetRandomName(0))
	}
	log.Println(names)

	watcher(a)

	// Initialize
	for _, name := range names {
		phils = append(phils, philosopher{name: name, announce: a, dying: d})
		forks = append(forks, fork(true))
	}
	for i := range phils {
		dude := &phils[i]
		dude.left = &(forks[i])
		if i == 0 {
			dude.right = &(forks[len(forks)-1])
		} else {
			dude.right = &(forks[i-1])
		}
	}

	for i := range phils {
		go phils[i].Live()
	}
	for i := 0; i < PHILOS; i++ {
		<-d
	}
	return
}

func main() {
	app := cli.NewApp()
	app.Name = "Philosophers Dinner experimentation"
	app.Usage = "Use cli flags to control testing environnement"
	app.Version = "1.0.2"

	app.Flags = []cli.Flag{
		cli.DurationFlag{
			Name:        "time, t",
			Usage:       "Time to run the experiment.",
			Destination: &duration,
			Value:       2 * time.Minute,
		},
		cli.IntFlag{
			Name:        "philo_number, n",
			Usage:       "How much dudes to simulate",
			Destination: &PHILOS,
			Value:       20,
		},
		cli.DurationFlag{
			Name:        "max-think-time, s",
			Usage:       "Maximum possible value for the *think* state.",
			Destination: &THINK_MAX_TIME,
			Value:       30 * time.Second,
		},
		cli.DurationFlag{
			Name:        "max-hungry-time, d",
			Usage:       "The limit of time a dude can be in *hungry* state before dying.",
			Destination: &HUNGRY_MAX_TIME,
			Value:       30 * time.Second,
		},
		cli.DurationFlag{
			Name:        "eat-time, e",
			Usage:       "The time it takes to eat",
			Destination: &EAT_TIME,
			Value:       10 * time.Second,
		},
	}

	app.Action = Run
	app.Run(os.Args)
}
