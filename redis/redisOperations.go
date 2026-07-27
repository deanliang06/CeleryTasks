package main

import "fmt"

func pollQueue() []byte {
	if queue.head == nil {
		return nil
	}

	node := queue.head
	queue.head = queue.head.next
	if queue.head == nil {
		queue.tail = nil
	}
	return node.data
}

func pushQueue(data []byte) []byte {
	newNode := LinkNode{
		data: data,
		next: nil,
	}
	if queue.head == nil {
		queue.head = &newNode
	}

	if queue.tail == nil {
		queue.tail = &newNode
	} else {
		queue.tail.next = &newNode
		queue.tail = &newNode
	}
	return nil
}

func addMap(data []byte) []byte {
	id, payload := parseMapPayload(data)
	fmt.Println(string(payload))
	taskMap.mut.Lock()
	taskMap.actMap[id] = payload
	taskMap.mut.Unlock()
	return nil
}

func getMap(data []byte) []byte {
	id, _ := parseMapPayload(data)
	fmt.Println(id)
	taskMap.mut.Lock()
	payload := taskMap.actMap[id]
	taskMap.mut.Unlock()
	return payload
}
