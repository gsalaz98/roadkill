package main

import "fmt"

type node struct {
	data int64
	next *node
}

func main() {
	obj := createNode(0)
	obj.insert(1)
	obj.insert(100)
	obj.insert((1 << 63) - 1)
	obj.insert(10000)

	obj.remove(0)
	obj.removeLast()

	tempObj := obj

	for tempObj.next != nil {
		fmt.Println(tempObj.data)
		tempObj = *tempObj.next
	}
	fmt.Println(tempObj.data)
}

func createNode(data int64) node {
	return node{
		data: data,
		next: nil,
	}
}

func (llist *node) insert(data int64) {
	var head = llist

	for (*head).next != nil {
		head = head.next
	}

	newNode := createNode(data)

	(*head).next = &newNode
}

func (llist *node) remove(index int) {
	var head = llist

	if index == 0 {
		*llist = *(llist.next)
		return
	}

	for i := 0; i < index; i++ {
		head = head.next
	}

	(*head).next = head.next.next
}

func (llist *node) removeFirst() {
	*llist = *(llist.next)
}

func (llist *node) removeLast() {
	var head = llist

	for head.next.next != nil {
		head = head.next
	}
	(*head).next = nil
}
