package g

import "time"

type Queue struct {
	length int
	data   []QueueComparer
}

type QueueComparer interface {
	Less(QueueComparer) bool
}

type QueueString struct {
	index  int64
	String string
}

func NewQueueString(s string) *QueueString {
	return &QueueString{index: time.Now().UnixNano(), String: s}
}

// Less Implementation QueueComparer interface.
func (a *QueueString) Less(b QueueComparer) bool {
	return a.index < b.(*QueueString).index
	//return strings.Compare(a.String, b.String) < 0
}

func NewQueue(data []QueueComparer) *Queue {
	q := &Queue{}
	q.data = data
	q.length = len(data)
	if q.length > 0 {
		i := q.length >> 1
		for ; i >= 0; i-- {
			q.down(i)
		}
	}
	return q
}

func (q *Queue) Push(item QueueComparer) {
	q.data = append(q.data, item)
	q.length++
	q.up(q.length - 1)
}

func (q *Queue) Pop() QueueComparer {
	if q.length == 0 {
		return nil
	}
	top := q.data[0]
	q.length--
	if q.length > 0 {
		q.data[0] = q.data[q.length]
		q.down(0)
	}
	q.data = q.data[:len(q.data)-1]
	return top
}

func (q *Queue) Peek() QueueComparer {
	if q.length == 0 {
		return nil
	}
	return q.data[0]
}

func (q *Queue) Len() int {
	return q.length
}

func (q *Queue) down(pos int) {
	data := q.data
	halfLength := q.length >> 1
	item := data[pos]
	for pos < halfLength {
		left := (pos << 1) + 1
		right := left + 1
		best := data[left]
		if right < q.length && data[right].Less(best) {
			left = right
			best = data[right]
		}
		if !best.Less(item) {
			break
		}
		data[pos] = best
		pos = left
	}
	data[pos] = item
}

func (q *Queue) up(pos int) {
	data := q.data
	item := data[pos]
	for pos > 0 {
		parent := (pos - 1) >> 1
		current := data[parent]
		if !item.Less(current) {
			break
		}
		data[pos] = current
		pos = parent
	}
	data[pos] = item
}
