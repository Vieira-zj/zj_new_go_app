package main

import "fmt"

// Spawner

type Spawner struct {
	queue chan Future
}

func (s Spawner) spawn(f Future) {
	go func() {
		s.queue <- f
	}()
}

// Executor

type Executor struct {
	num   int
	queue chan Future
}

func (e *Executor) incr() {
	e.num += 1
}

func (e *Executor) run() {
	for f := range e.queue {
		fmt.Printf("Executor: ready, and poll [%s]\n", f.string())
		if result := f.poll(); result == PollStatusReady {
			fmt.Printf("Executor: get [%s] ready\n", f.string())
			e.num -= 1
		}
		if e.isDone() {
			fmt.Println("Executor: all future done, and exit")
			return
		}
	}
}

func (e *Executor) isDone() bool {
	return e.num == 0
}
