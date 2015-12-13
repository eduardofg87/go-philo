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
	THINK_MAX_TIME    time.Duration
	EAT_TIME          time.Duration
	HUNGRY_MAX_TIME   time.Duration
	tempo             time.Duration
	PHILOS            int
	duration          time.Duration
	names             = []string{}
	verbose           bool
	important_methods = []string{"my life"}
)

type fork struct {
	toRight chan int
	toLeft  chan int
} //Free For future improvement

type announcement struct {
	from, message string
}

func is_important(method string) bool {
	for _, x := range important_methods {
		if x == method {
			return true
		}
	}
	return false
}

func (a announcement) String() string {
	//backstabbing_chandrasekhar
	return fmt.Sprintf("%-26s: %s", a.from, a.message)
}

type philosopher struct {
	name     string
	leftIn   <-chan int
	leftOut  chan<- int
	rightIn  <-chan int
	rightOut chan<- int
	state    string
	dying    chan int
	announce chan announcement
}

func (p *philosopher) timeTrack(from time.Time, method string) {
	if verbose || is_important(method) {
		p.say(fmt.Sprintf("I've finished %s in %vs. I'm now %s.", method, int64(time.Since(from))/1e9, p.state))
	}
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
				time.Sleep(time.Duration(rand.Intn(int(THINK_MAX_TIME))))
				p.state = "hungry"
			case "dead":
				return
			case "hungry":
				hunger := time.After(HUNGRY_MAX_TIME)
				for {
					select {
					case _ = <-p.leftIn:
						select {
						case _ = <-p.rightIn:
							p.state = "eat"
							return
						case _ = <-hunger:
							p.state = "dead"
							return

						}
					case _ = <-hunger:
						p.state = "dead"
						return
					}
					time.Sleep(tempo)
				}
				p.state = "dead"
			case "eat":
				time.Sleep(EAT_TIME)
				p.leftOut <- 1
				p.rightOut <- 1
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

func summarize() {
	fmt.Println("Summarize: Not implemented yet !")
	//TODO: Make a summarize function
}

func Run(c *cli.Context) {
	timeout := time.After(duration)
	a := make(chan announcement)
	defer timeTrack(time.Now(), "main", a)
	d := make(chan int)
	phils := []philosopher{}
	forks := []fork{}
	for i := PHILOS; i > 0; i-- {
		names = append(names, namesgenerator.GetRandomName(0))
	}
	log.Println(names)

	watcher(a)

	// Initialize
	for _, name := range names {
		phils = append(phils, philosopher{name: name, announce: a, dying: d})
		forks = append(forks, fork{toRight: make(chan int, 2), toLeft: make(chan int, 2)})
	}

	for i := range phils {
		dude := &phils[i]
		dude.leftIn, dude.leftOut = forks[i].toLeft, forks[i].toRight
		if i == 0 {
			dude.rightIn, dude.rightOut = forks[len(forks)-1].toRight, forks[len(forks)-1].toLeft
		} else {
			dude.rightIn, dude.rightOut = forks[i-1].toRight, forks[i-1].toLeft
		}
	}

	// Launch the dudes
	for i := range phils {
		go phils[i].Live()
	}
	//Put forks on the table
	for i, f := range forks {
		go func(i int, f fork) {
			if i%2 == 0 {
				f.toLeft <- 1
			} else {
				f.toRight <- 1
			}
		}(i, f)

	}
Wait:
	for i := 0; i < PHILOS; {
		select {
		case _ = <-d:
			i++
		case _ = <-timeout:
			a <- announcement{from: "Main", message: fmt.Sprintf("%v dudes died during the simulation, over the total of %v.", i, PHILOS)}
			break Wait
		}
	}
	summarize()
	return
}

func main() {
	rand.Seed(time.Now().Unix())
	app := cli.NewApp()
	app.Name = "Philosophers Dinner experimentation"
	app.Usage = "Use cli flags to control testing environnement"
	app.Version = "1.1.0"

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
			Value:       40 * time.Second,
		},
		cli.DurationFlag{
			Name:        "eat-time, e",
			Usage:       "The time it takes to eat",
			Destination: &EAT_TIME,
			Value:       10 * time.Second,
		},
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "More output",
			Destination: &verbose,
		},
	}

	app.Action = Run
	app.Run(os.Args)
}
