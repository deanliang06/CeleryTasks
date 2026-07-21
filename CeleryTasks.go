package main

import (
	"fmt"
	"sync"
	"time"
)

type Task struct {
	id int
	Func func() (any, error)
	tries int
	step string
}

type Celery struct {
	Wait chan *Task
	Stop chan struct{}
	NextID int	
	IDMut sync.Mutex
	wg sync.WaitGroup
	taskMap map[int]*Task
	workerWG sync.WaitGroup
}

func initCelery(numWorkers int) *Celery {
	wg := sync.WaitGroup{}

	c:=Celery{
		Wait: make(chan *Task, 100),
		Stop: make(chan struct{}), 
		NextID: 0,
		IDMut: sync.Mutex{},
		wg: wg,
		taskMap: make(map[int]*Task),
		workerWG: sync.WaitGroup{},
	}

	for i := range numWorkers {
		c.workerWG.Add(1)
		go c.initWorker(i)
	}

	return &c
} 

func (c *Celery) addTask(f func() (any, error)) (int, error) {
	c.IDMut.Lock()
	task := Task{
		id:c.NextID,
		Func: f,
		tries: 0,
		step: "waiting",
	}
	c.NextID++
	c.IDMut.Unlock()
	
	c.taskMap[task.id] = &task
	c.wg.Add(1)
	c.Wait <- &task
	return task.id, nil
}

func (c *Celery) initWorker(index int) {
	defer func() {
		c.workerWG.Done()
	}()
	for {
		select {
			case task:= <-c.Wait:
				if task.tries >= 3 {
					task.step="failed"
					msg:=fmt.Sprintf("Task with id %d failed max amount of times\n", task.id)
					fmt.Println(msg)
					c.wg.Done()
					continue
				}


				task.step = "running"
				res, e:= task.Func()
				msg:=""
				if e != nil {
					msg=fmt.Sprintf("Worker %d failed in execution of task with id %d with result: %v\n",  index, task.id, e.Error())
					task.step="waiting"
					task.tries++
					c.Wait<-task
					fmt.Println(msg)
				} else {
					msg=fmt.Sprintf("Worker %d completed %d with result: %v\n", index, task.id, res)
					task.step="completed"
					fmt.Println(msg)
					c.wg.Done()
				}
			case <-c.Stop:
				return 
		}
	}
}


func (c *Celery) termWorkers() {
	close(c.Stop)
}

func (c *Celery) reinitWorkers(numWorkers int) error {
	c.workerWG.Wait()
	c.Stop = make(chan struct{})

	for i := range numWorkers {
		go c.initWorker(i)
	}

	return nil
}


func main() {

	celeryQueue:=initCelery(8)
	
	timeStart:=time.Now()
	id1, _:=celeryQueue.addTask(
		func() (any, error) {
			piss(1)
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
			piss(2)
			return "Duh", nil
		},
	)

	id4, _:=celeryQueue.addTask(
		func() (any, error) {
			piss(1)
			return "Duh", nil
		},
	)

	fmt.Println(celeryQueue.taskMap[id1].step)
	fmt.Println(celeryQueue.taskMap[id4].step)


	
	celeryQueue.wg.Wait()
	fmt.Println(celeryQueue.taskMap[id4].step)

	celeryQueue.termWorkers()
	celeryQueue.reinitWorkers(8)

	celeryQueue.addTask(
		func() (any, error) {
			piss(3)
			return "Duh", nil
		},
	)
	celeryQueue.wg.Wait()

	fmt.Println("This took", time.Since(timeStart))
}

func piss(secs int) {
	time.Sleep(time.Second * time.Duration(secs))
}
