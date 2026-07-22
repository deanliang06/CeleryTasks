package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type healthForm struct {
	TaskType string `json:"taskType"`
	URL      string `json:"url"`
}

type TaskStatus struct {
	Id       int    `json:"id"`
	TaskType string `json:"taskType"`
	Status   string `json:"status"`
	Result   any    `json:"result"`
}

type TaskCreation struct {
	Id int `json:"id"`
}

func pollServer(taskID int) (TaskStatus, error) {
	res, err := http.Get("http://localhost:8000/getTask/" + strconv.Itoa(taskID))
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

func doTask(taskType string, url string) (any, error) {
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
	result, e := doTask("health", "https://pizzaposts.com/pizza/health")
	if e != nil {
		fmt.Println(e.Error())
	}
	fmt.Println(result)
}
