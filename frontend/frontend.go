package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type healthForm struct {
	TaskType string `json:"taskType"`
	URL      string `json:"url"`
}

type TaskStatus struct {
	Id       string `json:"id"`
	TaskType string `json:"taskType"`
	Status   string `json:"status"`
	Result   any    `json:"result"`
	Url      string `json:"url"`
}

type TaskCreation struct {
	Id string `json:"id"`
}

func pollServer(taskID string) (TaskStatus, error) {
	res, err := http.Get("http://localhost:8000/getTask/" + taskID)
	if err != nil {
		panic(err)
	}
	var status TaskStatus
	err = json.NewDecoder(res.Body).Decode(&status)
	res.Body.Close()
	if err != nil {
		return TaskStatus{}, fmt.Errorf("decode task status: %w", err)
	}
	return status, nil
}

func doTask(taskType, url string) (any, error) {
	payload := healthForm{
		TaskType: taskType,
		URL:      url,
	}

	body, e := json.Marshal(payload)
	if e != nil {
		return nil, e
	}
	res, err := http.Post("http://localhost:8000/addTask", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}

	var taskCreation TaskCreation
	err = json.NewDecoder(res.Body).Decode(&taskCreation)
	taskID := taskCreation.Id

	defer res.Body.Close()

	for {
		status, e := pollServer(taskID)
		if e != nil {
			panic(e)
		}

		if status.Status == "failed" {
			return nil, errors.New("Task of type %s failed")
		}

		if status.Status == "success" {
			return status.Result, nil
		}
		time.Sleep(time.Second * 5)
	}

}

func main() {
	uptime("https://pizzaposts.com/pizza/health", "get_health", "{\"detail\": \"healthy\"}")
}

func uptime(endpoint, taskType, healthyRes string) {
	var total, healthy float32
	healthy = 0
	total = 0
	for {
		result, e := doTask(taskType, endpoint)
		if e != nil {
			fmt.Println(e.Error())
		} else if result == healthyRes {
			healthy++
		}
		total++
		fmt.Printf("%s has uptime of %.2f%%\n", endpoint, healthy/total*100)
		time.Sleep(time.Second * 30)
	}
}
