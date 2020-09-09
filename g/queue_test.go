package g_test

import (
	"github.com/angenalZZZ/gofunc/g"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type floatValue float64

func (a floatValue) Less(b g.QueueComparer) bool {
	return a < b.(floatValue)
}

var queueData, queueSorted = func() ([]g.QueueComparer, []g.QueueComparer) {
	rand.Seed(time.Now().UnixNano())
	var data []g.QueueComparer
	for i := 0; i < 100; i++ {
		data = append(data, floatValue(rand.Float64()*100))
	}
	sorted := make([]g.QueueComparer, len(data))
	copy(sorted, data)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Less(sorted[j])
	})
	return data, sorted
}()

func TestMaintainsPriorityQueue(t *testing.T) {
	q := g.NewQueue(nil)
	for i := 0; i < len(queueData); i++ {
		q.Push(queueData[i])
	}
	assert.Equal(t, q.Peek(), queueSorted[0])
	var result []g.QueueComparer
	for q.Len() > 0 {
		result = append(result, q.Pop())
	}
	assert.Equal(t, result, queueSorted)
}

func TestAcceptsDataInConstructor(t *testing.T) {
	q := g.NewQueue(queueData)
	var result []g.QueueComparer
	for q.Len() > 0 {
		result = append(result, q.Pop())
	}
	assert.Equal(t, result, queueSorted)
}

func TestHandlesEdgeCasesWithFewElements(t *testing.T) {
	q := g.NewQueue(nil)
	q.Push(floatValue(2))
	q.Push(floatValue(1))
	q.Pop()
	q.Pop()
	q.Pop()
	q.Push(floatValue(2))
	q.Push(floatValue(1))
	assert.Equal(t, float64(q.Pop().(floatValue)), 1.0)
	assert.Equal(t, float64(q.Pop().(floatValue)), 2.0)
	assert.Equal(t, q.Pop(), nil)
}
