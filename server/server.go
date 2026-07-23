package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type healthForm struct {
	Id       string `json:"id"`
	TaskType string `json:"taskType"`
	URL      string `json:"url"`
	Tries    int    `json:"tries"`
}

type mapEntry struct {
	Id       string `json:"id"`
	TaskType string `json:"taskType"`
	Status   string `json:"status"`
	Result   any    `json:"result"`
	Url      string `json:"url"`
}

func addTask(w http.ResponseWriter, r *http.Request) {
	var req healthForm

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Println("error", err)
	}

	if req.TaskType == "get_health" {
		client := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
			Protocol: 2,
		})

		ctx := context.Background()

		IdToUse := uuid.New().String()
		mapEntry := mapEntry{
			Id:       IdToUse,
			TaskType: req.TaskType,
			Status:   "queued",
			Result:   "N/A",
			Url:      req.URL,
		}

		marshalledMap, err := json.Marshal(mapEntry)
		if err != nil {
			panic(err)
		}

		err = client.HSet(ctx, "taskMap", IdToUse, marshalledMap).Err()
		if err != nil {
			fmt.Println("Redis connection is problemo")
			panic(err)
		}
		req.Id = IdToUse
		req.Tries = 0

		marshalledTask, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}

		err = client.LPush(ctx, "taskQueue:toBe", marshalledTask).Err()
		if err != nil {
			panic(err)
		}

		unmarshaled, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}
		w.Write(unmarshaled)
	}
}

func getTaskInfo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("taskId")
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	ctx := context.Background()
	boolRes, err := client.HExists(ctx, "taskMap", id).Result()
	if !boolRes {
		panic("Fuck you the key doesn't exist")
	}
	res, err := client.HGet(ctx, "taskMap", id).Result()
	if err != nil {
		panic(err)
	}

	var entry mapEntry
	json.NewDecoder(bytes.NewReader([]byte(res))).Decode(&entry)

	if entry.TaskType == "get_health" {
		json.NewEncoder(w).Encode(entry)
		fmt.Println("Ping")
	}

}

// responsible for handling task requests and adding to redis queue
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /addTask", addTask)
	mux.HandleFunc("GET /getTask/{taskId}", getTaskInfo)

	err := http.ListenAndServe(":8000", mux)
	if err != nil {
		fmt.Println("errors shit", err.Error())
	}
}
