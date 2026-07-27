package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"encoding/json"
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

func initCelery(numWorkers int) {

	for i := range numWorkers {
		go initWorker(i)
	}
}

func initWorker(index int) {
	client, err := initConn("localhost:8080")
	if err != nil {
		panic(err)
	}
	for {
		task, err := client.waitQueue()
		if err != nil {
			panic(err)
		}

		var unMarshTask QueueTask
		json.Unmarshal(task, &unMarshTask)

		var taskInfo MapEntry
		entry, err := client.getMap(unMarshTask.Id)
		if err != nil {
			panic(err)
		}

		json.Unmarshal([]byte(entry), &taskInfo)
		newTries := taskInfo.Tries + 1

		changeMap(client, taskInfo.Id, "executing", newTries)

		if taskInfo.TaskType != "get_health" {
			panic(errors.New("Not accepted task type"))
		}
		res, e := getPage(taskInfo.Url)

		if e != nil {
			msg := fmt.Sprintf("Worker %d failed in execution of task with id %d with result: %v\n", index, taskInfo.Id, e.Error())
			if taskInfo.Tries+1 >= maxTries {
				changeMap(client, taskInfo.Id, "failed", newTries)
			} else {
				changeMap(client, taskInfo.Id, "queued", newTries)
				err := reinsertQueue(client, unMarshTask)
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
			changeMap(client, taskInfo.Id, "success", newTries)
			setMapResult(client, taskInfo.Id, stringRes)
			fmt.Println(msg)
		}
	}
	client.Close()
}

func setMapResult(client *redisConn, id string, result string) {
	res, err := client.getMap(id)
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

	err = client.addMap(id, byteEntry)
	if err != nil {
		panic(err)
	}
}

func changeMap(client *redisConn, id string, status string, tries int) {
	res, err := client.getMap(id)
	if err != nil {
		panic(err)
	}

	var entry MapEntry
	json.NewDecoder(bytes.NewReader(res)).Decode(&entry)

	entry.Status = status
	entry.Tries = tries

	marshalled, err := json.Marshal(entry)
	if err != nil {
		panic(err)
	}
	err = client.addMap(id, marshalled)
	if err != nil {
		panic(err)
	}
}
func reinsertQueue(client *redisConn, task QueueTask) error {
	marsh, err := json.Marshal(task)
	if err != nil {
		panic(err)
	}

	client.pushQueue(marsh)
	return nil
}

// responsible for initiating workers which pull from redis queues
func main() {
	initCelery(8)
	select {}
}

func getPage(url string) (*http.Response, error) {
	return http.Get(url)
}
