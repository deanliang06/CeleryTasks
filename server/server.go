package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
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
		client, err := initConn("localhost:8080")
		if err != nil {
			panic(err)
		}

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

		err = client.addMap(IdToUse, marshalledMap)
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
		err = client.pushQueue(marshalledTask)
		if err != nil {
			panic(err)
		}

		unmarshaled, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}
		w.Write(unmarshaled)
		client.Close()
	}
}

func getTaskInfo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("taskId")
	client, err := initConn("localhost:8080")
	if err != nil {
		panic(err)
	}

	res, err := client.getMap(id)
	if err != nil {
		panic(err)
	}

	var entry mapEntry
	json.NewDecoder(bytes.NewReader([]byte(res))).Decode(&entry)

	if entry.TaskType == "get_health" {
		json.NewEncoder(w).Encode(entry)
	}
	client.Close()
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
	// testRedis()
}

// func testRedis() {
// 	conn, err := initConn("localhost:8080")
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	conn.addMap("car1", []byte("Green and sassy"))
// 	conn.addMap("car2", []byte("BLue and mellow"))
// 	conn.addMap("car3", []byte("Red and sascwatch"))

// 	output, err := conn.getMap("car1")
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	fmt.Println(string(output))

// 	output, err = conn.getMap("car3")
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	fmt.Println(string(output))

// 	conn.Close()
// }
