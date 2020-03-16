package queue

import (
	"os"
	"path/filepath"
)

// queueType defines the type of queue data structure used.
type queueType uint8

// The possible queue types, used to determine compatibility when
// one stored type is trying to be opened by a different type.
const (
	queueStack queueType = iota
	queueQueue
	queuePriorityQueue
	queuePrefixQueue
)

// checkQueueType checks if the type of queue data structure
// trying to be opened is compatible with the opener type.
//
// A file named 'queue' within the data directory used by
// the structure stores the structure type, using the constants
// declared above.
//
// Stacks and Queues are 100% compatible with each other, while
// a PriorityQueue is incompatible with both.
//
// Returns true if types are compatible and false if incompatible.
func checkQueueType(dataDir string, gt queueType) (bool, error) {
	// Set the path to 'queue' file.
	path := filepath.Join(dataDir, "queue")

	// Read 'queue' file for this directory.
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if os.IsNotExist(err) {
		f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return false, err
		}
		defer f.Close()

		// Create byte slice of queueType.
		gtb := make([]byte, 1)
		gtb[0] = byte(gt)

		_, err = f.Write(gtb)
		if err != nil {
			return false, err
		}

		return true, nil
	}
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Get the saved type from the file.
	fb := make([]byte, 1)
	_, err = f.Read(fb)
	if err != nil {
		return false, err
	}

	// Convert the file byte to its queueType.
	filegt := queueType(fb[0])

	// Compare the types.
	if filegt == gt {
		return true, nil
	} else if filegt == queueStack && gt == queueQueue {
		return true, nil
	} else if filegt == queueQueue && gt == queueStack {
		return true, nil
	}

	return false, nil
}
