package main

import (
	"fmt"
	"sync"
	"time"
)

type Task struct {
	id int
	Func func() (any, error)
}

type Celery struct {
	Wait chan Task
	NextID int	
	IDMut sync.Mutex
	wg sync.WaitGroup
}

func initCelery(numWorkers int) *Celery {
	wg := sync.WaitGroup{}

	c:=Celery{
		Wait: make(chan Task, 100),
		NextID: 0,
		IDMut: sync.Mutex{},
		wg: wg,
	}

	for i := range numWorkers {
		go c.initWorker(i)
	}

	return &c
} 

func (c *Celery) addTask(f func() (any, error)) (string, error) {
	c.IDMut.Lock()
	task := Task{
		id:c.NextID,
		Func: f,
	}
	c.NextID++
	c.IDMut.Unlock()

	c.Wait <- task
	c.wg.Add(1)
	return "Added", nil
}

func (c *Celery) initWorker(index int) {
	for task := range c.Wait {
		res, e:=task.Func()
		msg:=""
		if e != nil {
			msg=fmt.Sprintf("Worker %d failed in execution %d with result: %v\n",  index, task.id, e.Error())
		} else {
			msg=fmt.Sprintf("Worker %d completed %d with result: %v\n", index, task.id, res)
		}

		fmt.Println(msg)
		c.wg.Done()
	}
}


func main() {

	celeryQueue:=initCelery(8)
	
	timeStart:=time.Now()
	celeryQueue.addTask(
		func() (any, error) {
			piss(2)
			return "Bob", nil
		},
	)
	celeryQueue.addTask(
		func() (any, error) {
			piss(3)
			return "Duh", nil
		},
	)

	celeryQueue.addTask(
		func() (any, error) {
			piss(4)
			return "Duh", nil
		},
	)

	celeryQueue.addTask(
		func() (any, error) {
			piss(3)
			return "Duh", nil
		},
	)
	
	celeryQueue.wg.Wait()

	fmt.Println("This took %v secs", time.Since(timeStart))
}

func piss(secs int64) {
	time.Sleep(time.Second * time.Duration(secs))
}
