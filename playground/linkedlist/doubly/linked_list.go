package main

import "fmt"

type node struct {
	data int64
	prev *node
	next *node
}

func main() {
	a := createNode(0)
	a.insertTail(20)
	a.insertTail(40)
	a.insertHead(-20)

	var temp = a

	for temp.next != nil {
		fmt.Println(temp.data)
		temp = *temp.next
	}
	for temp.prev != nil {
		fmt.Println(temp.data, temp.prev)
		temp = *temp.prev
	}
	fmt.Println(temp.data)
}

func createNode(data int64) node {
	return node{
		data: data,
		prev: nil,
		next: nil,
	}
}

func (llist *node) insertTail(data int64) {
	temp := llist

	for temp.next != nil {
		temp = temp.next
	}
	newNode := createNode(data)
	(*temp).next = &newNode
	(*temp).next.prev = &(*temp)
}

func (llist *node) insertHead(data int64) {
	head := createNode(data)
	llist.prev = &head
	llistCopy := *llist
	head.next = &llistCopy

	*llist = head
}
