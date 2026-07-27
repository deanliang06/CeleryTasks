package main

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
	}
	return nil
}
