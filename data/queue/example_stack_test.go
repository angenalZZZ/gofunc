package queue_test

import (
	"fmt"

	"github.com/angenalZZZ/gofunc/data/queue"
)

// ExampleStack demonstrates the implementation of a queue stack.
func Example_stack() {
	// Open/create a stack.
	s, err := queue.OpenStack("data_dir")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Close()

	// Push an item onto the stack.
	item, err := s.Push([]byte("item value"))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(item.ID)         // 1
	fmt.Println(item.Key)        // [0 0 0 0 0 0 0 1]
	fmt.Println(item.Value)      // [105 116 101 109 32 118 97 108 117 101]
	fmt.Println(item.ToString()) // item value

	// Change the item value in the stack.
	item, err = s.Update(item.ID, []byte("new item value"))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(item.ToString()) // new item value

	// Pop an item off the stack.
	popItem, err := s.Pop()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(popItem.ToString()) // new item value

	// Delete the stack and its database.
	s.Drop()
}
