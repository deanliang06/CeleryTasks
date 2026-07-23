package main

import (
	"context"
	"net/http"
	"time"

	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type QueueTask struct {
	Id       string `json:"id"`
	TaskType string `json:"taskType"`
	URL      string `json:"url"`
	Tries    int    `json:"tries"`
}

// type Celery struct {
// 	// Stop chan struct{}
// 	// workerWG sync.WaitGroup
// }

func initCelery(numWorkers int) {
	// c := Celery{
	// 	// Stop: make(chan struct{}),
	// 	// workerWG: sync.WaitGroup{},
	// }

	for i := range numWorkers {
		// c.workerWG.Add(1)
		go initWorker(i)
	}
}

func initWorker(index int) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	for {
		task, err := client.BRPop(ctx, time.Hour, "taskQueue:toBe").Result()
		if err != nil {
			panic(err)
		}

		var unMarshTask QueueTask
		json.Unmarshal([]byte(task[1]), &unMarshTask)

		// task.step = "running"
		// res, e := task.Func()
		// msg := ""
		// if e != nil {
		// 	msg = fmt.Sprintf("Worker %d failed in execution of task with id %d with result: %v\n", index, task.id, e.Error())
		// 	task.step = "waiting"
		// 	task.tries++

		// 	fmt.Println(msg)
		// } else {
		// 	msg = fmt.Sprintf("Worker %d completed %d with result: %v\n", index, task.id, res)
		// 	task.step = "completed"
		// 	fmt.Println(msg)
		// }
	}
}

// func reinsertQueue(task QueueTask) error {

// }

// func (c *Celery) termWorkers() {
// 	close(c.Stop)
// }

// func (c *Celery) reinitWorkers(numWorkers int) error {
// 	c.workerWG.Wait()
// 	c.Stop = make(chan struct{})

// 	for i := range numWorkers {
// 		go c.initWorker(i)
// 	}

// 	return nil
// }

// responsible for initiating workers which pull from redis queues
func main() {
	initCelery(8)
	select {}
}

func getPage(url string) (any, error) {
	return http.Get(url)
}
