package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"encoding/json"

	"github.com/redis/go-redis/v9"
)

var maxTries int = 3

type QueueTask struct {
	Id       string `json:"id"`
	TaskType string `json:"taskType"`
	URL      string `json:"url"`
}

type MapEntry struct {
	Id       string `json:"id"`
	TaskType string `json:"taskType"`
	Status   string `json:"status"`
	Result   any    `json:"result"`
	Url      string `json:"url"`
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

		var taskInfo MapEntry
		entry, err := client.HGet(ctx, "taskMap", unMarshTask.Id).Result()
		if err != nil {
			panic(err)
		}

		json.Unmarshal([]byte(entry), &taskInfo)
		newTries := taskInfo.Tries + 1

		changeMap(ctx, client, taskInfo.Id, "executing", newTries)

		if taskInfo.TaskType != "get_health" {
			panic(errors.New("Not accepted task type"))
		}
		res, e := getPage(taskInfo.Url)

		if e != nil {
			msg := fmt.Sprintf("Worker %d failed in execution of task with id %d with result: %v\n", index, taskInfo.Id, e.Error())
			if taskInfo.Tries+1 >= maxTries {
				changeMap(ctx, client, taskInfo.Id, "failed", newTries)
			} else {
				changeMap(ctx, client, taskInfo.Id, "queued", newTries)
				err := reinsertQueue(ctx, client, unMarshTask)
				if err != nil {
					panic(err)
				}
			}
			fmt.Println(msg)
		} else {
			byteRes, err := io.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}

			stringRes := string(byteRes)
			res.Body.Close()

			msg := fmt.Sprintf("Worker %d completed task with id %s with result: %v\n", index, taskInfo.Id, stringRes)
			changeMap(ctx, client, taskInfo.Id, "success", newTries)
			setMapResult(ctx, client, taskInfo.Id, stringRes)
			fmt.Println(msg)
		}
	}
}

func setMapResult(ctx context.Context, client *redis.Client, id string, result string) {
	res, err := client.HGet(ctx, "taskMap", id).Result()
	if err != nil {
		panic(err)
	}

	var entry MapEntry
	json.Unmarshal([]byte(res), &entry)
	entry.Result = result
	byteEntry, err := json.Marshal(entry)
	if err != nil {
		panic(err)
	}

	err = client.HSet(ctx, "taskMap", id, byteEntry).Err()
	if err != nil {
		panic(err)
	}
}

func changeMap(ctx context.Context, client *redis.Client, id string, status string, tries int) {
	res, err := client.HGet(ctx, "taskMap", id).Result()
	if err != nil {
		panic(err)
	}

	var entry MapEntry
	json.NewDecoder(bytes.NewReader([]byte(res))).Decode(&entry)

	entry.Status = status
	entry.Tries = tries

	marshalled, err := json.Marshal(entry)
	if err != nil {
		panic(err)
	}
	err = client.HSet(ctx, "taskMap", id, marshalled).Err()
	if err != nil {
		panic(err)
	}
}
func reinsertQueue(ctx context.Context, client *redis.Client, task QueueTask) error {
	marsh, err := json.Marshal(task)
	if err != nil {
		panic(err)
	}

	client.LPush(ctx, "taskQueue:toBe", marsh)
	return nil
}

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

func getPage(url string) (*http.Response, error) {
	return http.Get(url)
}
