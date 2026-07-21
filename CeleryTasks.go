package main

import (
	"fmt"
	"time"
	"sync"
)

type Result struct {
	ID int
	ActResult any
}

type Task struct {
	id int
	Func func() any
}

type Celery struct {
	Result chan Result
	Wait chan Task
	Workers int
	NextID int
}

func (c *Celery) addTask(f func() any) (string, error) {
	task := Task{
		id:c.NextID,
		Func: f,
	}

	c.Wait <- task
	go c.tryExecute()

	c.NextID++

	return "Added", nil
}

func (c *Celery) tryExecute() (string, error) {
	if c.Workers > 0 {
		c.Workers--
		task := <-c.c
		c.Result <- Result{
			ID: task.id,
			ActResult: task.Func(),
		}
		c.Workers++
		return "Let's go", nil
	} else {
		return "No more workers", nil
	}
}

func (c *Celery) sync() {
	for c.Workers != 8 {
		time.Sleep(time.Second)
	}

	close(c.Result)

	for result := range c.Result {
		fmt.Println(result.ActResult)
	}
}


func main() {

	celeryQueue:=Celery{
		Wait:make(chan Task, 100),
		Workers:8,
		NextID:1,
	}
	
	celeryQueue.addTask(
		func() any {
			dog()
			return "Bob"
		},
	)
	celeryQueue.addTask(
		func() any {
			print("cat")
			return "Duh"
		},
	)

	celeryQueue.sync()
}

func dog() {
	fmt.Println("dog")
}

func print(s string) {
	fmt.Println(s)
}

