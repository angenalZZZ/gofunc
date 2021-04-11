package snowid

import (
	"sync"
)

var singletonMutex sync.Mutex
var idGenerator *DefaultIdGenerator

// SetDefaultIdGenerator Set default options.
func SetDefaultIdGenerator(options *IdGeneratorOptions) {
	singletonMutex.Lock()
	idGenerator = NewDefaultIdGenerator(options)
	singletonMutex.Unlock()
}

// NextId Create a new ID.
func NextId() uint64 {
	if idGenerator == nil {
		singletonMutex.Lock()
		defer singletonMutex.Unlock()
		options := NewIdGeneratorOptions(1)
		idGenerator = NewDefaultIdGenerator(options)
	}

	return idGenerator.NewLong()
}
