package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type healthForm struct {
	TaskType string `json:"taskType"`
	URL      string `json:"url"`
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
		client.LPush(ctx, "taskQueue:toBe", req)
	}
}

// responsible for handling task requests and adding to redis queue
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /addTask", addTask)

	http.ListenAndServe(":8000", nil)
}
