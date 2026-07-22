package main

import (
	"fmt"
	"sync"
	"errors"
	"net/http"
	"github.com/redis/go-redis/v9"
	"context"
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

func (c *Celery) addTask(taskType string, params ...any) (int, error) {
	var f func() (any, error)
	if taskType == "get_health" {
		if (len(params) > 1 || len(params) < 1) {
			return -1, errors.New("You shit the bed lil bro")
		}
		f=func() (any, error) {return getPage(params[0].(string))}
	}
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

//responsible for initiating workers which pull from redis queues
func main() {
	client:= redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	ctx := context.Background()

	res15, err := client.RPush(ctx, "bikes:repairs", "bike:1").Result()

	if err != nil {
		panic(err)
	}

	fmt.Println(res15)


	celeryQueue:=initCelery(8)
	fmt.Println(celeryQueue)
	
}

func getPage(url string) (any, error) {
	return http.Get(url)
}
