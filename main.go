package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"os/exec"
)

var (
	THINK_MAX_TIME int = 8
	EAT_TIME time.Duration = 10 * time.Second
	HUNGRY_MAX_TIME time.Duration = 3 * EAT_TIME
	PHILOS = 300
	names = []string{}
)

type fourchette bool // Used/Free For future improvement

type announcement struct {
	from, message string
}

func (a announcement) String() string {
	return fmt.Sprintf("%s: %s", a.from, a.message)
}

type philosopher struct {
	name     string
	left     *fourchette
	right    *fourchette
	state    *string
	dying    chan int
	announce chan announcement
}

func (p philosopher) timeTrack(from time.Time, method string) {
	p.say(fmt.Sprintf("I've finished %s in %vs. I'm now %s.", method, int64(time.Since(from)) / 1e9, *p.state))
}

func (p philosopher) say(message string) {
	p.announce <- announcement{from: p.name, message: message}
}

func (p philosopher) Live() {
	p.state = new(string)


	defer p.timeTrack(time.Now(), "my life")
	defer func() {
		*p.state = "dead"
		p.dying <- 1
	}()
	*p.state = "hungry"
	for *p.state != "dead" {
		from := time.Now()
		func() {
			defer p.timeTrack(from, *p.state)
			switch *p.state {
			case "think":
				time.Sleep(time.Duration(rand.Intn(THINK_MAX_TIME) + 2) * time.Second)
				*p.state = "hungry"
			case "dead":
				return
			case "hungry":
				for time.Since(from) < HUNGRY_MAX_TIME {
					if *p.left && *p.right {
						*p.left, *p.right = false, false
						*p.state = "eat"
						return
					}
					time.Sleep(200 * time.Millisecond)
				}
				*p.state = "dead"
			case "eat":
				time.Sleep(EAT_TIME)
				*p.left = true
				*p.right = true
				*p.state = "think"
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

func main() {
	a := make(chan announcement)
	defer timeTrack(time.Now(), "main", a)
	d := make(chan int)
	phils := []philosopher{}
	forks := []fourchette{}
	for i := PHILOS; i >= 0; i-- {
		out, err := exec.Command("uuidgen").Output()
		if err != nil {
			log.Fatal(err)
		}
		names = append(names, string(out[:len(out) - 1]))
	}
	log.Println(names)

	watcher(a)

	// Initialize
	for _, name := range names {
		phils = append(phils, philosopher{name: name, announce: a, dying: d})
		forks = append(forks, fourchette(true))
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
		_ = <-d
	}
	return
}
