package f

import (
	"container/heap"
	"container/list"
	"errors"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Timer is an implementation of Hierarchical Timing Wheels.
type Timer struct {
	tick      int64 // in milliseconds
	wheelSize int64

	interval    int64 // in milliseconds
	currentTime int64 // in milliseconds
	buckets     []*TimerBucket
	queue       *timerDelayQueue

	// The higher-level overflow wheel.
	//
	// NOTE: This field may be updated and read concurrently, through Add().
	overflowWheel unsafe.Pointer // type: *Timer

	exitC     chan struct{}
	waitGroup sync.WaitGroup
}

// NewTimer creates an instance of Timer with the given tick and wheelSize.
func NewTimer(tick time.Duration, wheelSize int64) *Timer {
	tickMs := int64(tick / time.Millisecond)
	if tickMs <= 0 {
		panic(errors.New("tick must be greater than or equal to 1ms"))
	}

	startMs := time.Now().UTC().UnixNano() / int64(time.Millisecond)

	return newTimer(
		tickMs,
		wheelSize,
		startMs,
		newTimerDelayQueue(int(wheelSize)),
	)
}

// newTimer is an internal helper function that really creates an instance of Timer.
func newTimer(tickMs int64, wheelSize int64, startMs int64, queue *timerDelayQueue) *Timer {
	currentTime := startMs
	if tickMs > 0 {
		currentTime = startMs - startMs%tickMs
	}
	buckets := make([]*TimerBucket, wheelSize)
	for i := range buckets {
		buckets[i] = NewTimerBucket()
	}
	return &Timer{
		tick:        tickMs,
		wheelSize:   wheelSize,
		currentTime: currentTime,
		interval:    tickMs * wheelSize,
		buckets:     buckets,
		queue:       queue,
		exitC:       make(chan struct{}),
	}
}

// add inserts the timer t into the current timing wheel.
func (tw *Timer) add(t *TimerElement) bool {
	currentTime := atomic.LoadInt64(&tw.currentTime)
	if t.expiration < currentTime+tw.tick {
		// Already expired
		return false
	} else if t.expiration < currentTime+tw.interval {
		// Put it into its own bucket
		virtualID := t.expiration / tw.tick
		b := tw.buckets[virtualID%tw.wheelSize]
		b.Add(t)

		// Set the bucket expiration time
		if b.SetExpiration(virtualID * tw.tick) {
			// The bucket needs to be enqueued since it was an expired bucket.
			// We only need to enqueue the bucket when its expiration time has changed,
			// i.e. the wheel has advanced and this bucket get reused with a new expiration.
			// Any further calls to set the expiration within the same wheel cycle will
			// pass in the same value and hence return false, thus the bucket with the
			// same expiration will not be enqueued multiple times.
			tw.queue.Offer(b, b.Expiration())
		}
		return true
	} else {
		// Out of the interval. Put it into the overflow wheel
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel == nil {
			atomic.CompareAndSwapPointer(
				&tw.overflowWheel,
				nil,
				unsafe.Pointer(newTimer(
					tw.interval,
					tw.wheelSize,
					currentTime,
					tw.queue,
				)),
			)
			overflowWheel = atomic.LoadPointer(&tw.overflowWheel)
		}
		return (*Timer)(overflowWheel).add(t)
	}
}

// addOrRun inserts the timer t into the current timing wheel, or run the
// timer's task if it has already expired.
func (tw *Timer) addOrRun(t *TimerElement) {
	if !tw.add(t) {
		// Already expired

		// Like the standard time.AfterFunc (https://golang.org/pkg/time/#AfterFunc),
		// always execute the timer's task in its own goroutine.
		go t.task()
	}
}

func (tw *Timer) advanceClock(expiration int64) {
	currentTime := atomic.LoadInt64(&tw.currentTime)
	if expiration >= currentTime+tw.tick {
		currentTime := expiration
		if tw.tick > 0 {
			currentTime = expiration - expiration%tw.tick
		}
		atomic.StoreInt64(&tw.currentTime, currentTime)

		// Try to advance the clock of the overflow wheel if present
		overflowWheel := atomic.LoadPointer(&tw.overflowWheel)
		if overflowWheel != nil {
			(*Timer)(overflowWheel).advanceClock(currentTime)
		}
	}
}

// Start starts the current timing wheel.
func (tw *Timer) Start() {
	tw.waitGroup.Add(1)
	go func() {
		defer tw.waitGroup.Done()
		tw.queue.Poll(tw.exitC, func() int64 {
			return time.Now().UTC().UnixNano() / int64(time.Millisecond)
		})
	}()

	tw.waitGroup.Add(1)
	go func() {
		defer tw.waitGroup.Done()
		for {
			select {
			case elem := <-tw.queue.C:
				b := elem.(*TimerBucket)
				tw.advanceClock(b.Expiration())
				b.Flush(tw.addOrRun)
			case <-tw.exitC:
				return
			}
		}
	}()
}

// Stop stops the current timing wheel.
//
// If there is any timer's task being running in its own goroutine, Stop does
// not wait for the task to complete before returning. If the caller needs to
// know whether the task is completed, it must coordinate with the task explicitly.
func (tw *Timer) Stop() {
	close(tw.exitC)
	tw.waitGroup.Wait()
}

// AfterFunc waits for the duration to elapse and then calls f in its own goroutine.
// It returns a Timer that can be used to cancel the call using its Stop method.
func (tw *Timer) AfterFunc(d time.Duration, f func()) *TimerElement {
	t := &TimerElement{
		expiration: time.Now().UTC().Add(d).UnixNano() / int64(time.Millisecond),
		task:       f,
	}
	tw.addOrRun(t)
	return t
}

// TimerScheduler determines the execution plan of a task.
type TimerScheduler interface {
	// Next returns the next execution time after the given (previous) time.
	// It will return a zero time if no next time is scheduled.
	//
	// All times must be UTC.
	Next(time.Time) time.Time
}

// ScheduleFunc calls f (in its own goroutine) according to the execution
// plan scheduled by s. It returns a Timer that can be used to cancel the
// call using its Stop method.
//
// If the caller want to terminate the execution plan halfway, it must
// stop the timer and ensure that the timer is stopped actually, since in
// the current implementation, there is a gap between the expiring and the
// restarting of the timer. The wait time for ensuring is short since the
// gap is very small.
//
// Internally, ScheduleFunc will ask the first execution time (by calling
// s.Next()) initially, and create a timer if the execution time is non-zero.
// Afterwards, it will ask the next execution time each time f is about to
// be executed, and f will be called at the next execution time if the time
// is non-zero.
func (tw *Timer) ScheduleFunc(s TimerScheduler, f func()) (t *TimerElement) {
	expiration := s.Next(time.Now().UTC())
	if expiration.IsZero() {
		// No time is scheduled, return nil.
		return
	}

	t = &TimerElement{
		expiration: expiration.UnixNano() / int64(time.Millisecond),
		task: func() {
			// Schedule the task to execute at the next time if possible.
			expiration := s.Next(time.Unix(0, t.expiration*int64(time.Millisecond)).UTC())
			if !expiration.IsZero() {
				t.expiration = expiration.UnixNano() / int64(time.Millisecond)
				tw.addOrRun(t)
			}

			// Actually execute the task.
			f()
		},
	}
	tw.addOrRun(t)

	return
}

// The start of TimerElement implementation.

// TimerElement represents a single event. When the Timer expires, the given
// task will be executed.
type TimerElement struct {
	expiration int64 // in milliseconds
	task       func()

	// The bucket that holds the list to which this timer's element belongs.
	//
	// NOTE: This field may be updated and read concurrently,
	// through Timer.Stop() and Bucket.Flush().
	b unsafe.Pointer // type: *bucket

	// The timer's element.
	element *list.Element
}

func (t *TimerElement) getBucket() *TimerBucket {
	return (*TimerBucket)(atomic.LoadPointer(&t.b))
}

func (t *TimerElement) setBucket(b *TimerBucket) {
	atomic.StorePointer(&t.b, unsafe.Pointer(b))
}

// Stop prevents the Timer from firing. It returns true if the call
// stops the timer, false if the timer has already expired or been stopped.
//
// If the timer t has already expired and the t.task has been started in its own
// goroutine; Stop does not wait for t.task to complete before returning. If the caller
// needs to know whether t.task is completed, it must coordinate with t.task explicitly.
func (t *TimerElement) Stop() bool {
	stopped := false
	for b := t.getBucket(); b != nil; b = t.getBucket() {
		// If b.Remove is called just after the timing wheel's goroutine has:
		//     1. removed t from b (through b.Flush -> b.remove)
		//     2. moved t from b to another bucket ab (through b.Flush -> b.remove and ab.Add)
		// this may fail to remove t due to the change of t's bucket.
		stopped = b.Remove(t)

		// Thus, here we re-get t's possibly new bucket (nil for case 1, or ab (non-nil) for case 2),
		// and retry until the bucket becomes nil, which indicates that t has finally been removed.
	}
	return stopped
}

// TimerBucket manage timers list
type TimerBucket struct {
	// 64-bit atomic operations require 64-bit alignment, but 32-bit
	// compilers do not ensure it. So we must keep the 64-bit field
	// as the first field of the struct.
	//
	// For more explanations, see https://golang.org/pkg/sync/atomic/#pkg-note-BUG
	expiration int64

	mu     sync.Mutex
	Timers *list.List
}

// NewTimerBucket creates timers list
func NewTimerBucket() *TimerBucket {
	return &TimerBucket{
		Timers:     list.New(),
		expiration: -1,
	}
}

// Expiration get expiration time
func (b *TimerBucket) Expiration() int64 {
	return atomic.LoadInt64(&b.expiration)
}

// SetExpiration set expiration time
func (b *TimerBucket) SetExpiration(expiration int64) bool {
	return atomic.SwapInt64(&b.expiration, expiration) != expiration
}

// Add a element to timers list
func (b *TimerBucket) Add(t *TimerElement) {
	b.mu.Lock()

	e := b.Timers.PushBack(t)
	t.setBucket(b)
	t.element = e

	b.mu.Unlock()
}

// Remove a element to timers list
func (b *TimerBucket) Remove(t *TimerElement) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.remove(t)
}

func (b *TimerBucket) remove(t *TimerElement) bool {
	if t.getBucket() != b {
		// If remove is called from t.Stop, and this happens just after the timing wheel's goroutine has:
		//     1. removed t from b (through b.Flush -> b.remove)
		//     2. moved t from b to another bucket ab (through b.Flush -> b.remove and ab.Add)
		// then t.getBucket will return nil for case 1, or ab (non-nil) for case 2.
		// In either case, the returned value does not equal to b.
		return false
	}
	b.Timers.Remove(t.element)
	t.setBucket(nil)
	t.element = nil
	return true
}

// Flush timers list
func (b *TimerBucket) Flush(reinsert func(*TimerElement)) {
	var ts []*TimerElement

	b.mu.Lock()
	for e := b.Timers.Front(); e != nil; {
		next := e.Next()

		t := e.Value.(*TimerElement)
		b.remove(t)
		ts = append(ts, t)

		e = next
	}
	b.mu.Unlock()

	b.SetExpiration(-1)

	for _, t := range ts {
		reinsert(t)
	}
}

// The end of TimerElement implementation.

// The start of PriorityQueue implementation.
// Borrowed from https://github.com/nsqio/nsq/blob/master/internal/pqueue/pqueue.go

type priorityQueueItem struct {
	Value    interface{}
	Priority int64
	Index    int
}

// this is a priority queue as implemented by a min heap
// ie. the 0th element is the *lowest* value
type priorityQueue []*priorityQueueItem

func newPriorityQueue(capacity int) priorityQueue {
	return make(priorityQueue, 0, capacity)
}

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	c := cap(*pq)
	if n+1 > c {
		npq := make(priorityQueue, n, c*2)
		copy(npq, *pq)
		*pq = npq
	}
	*pq = (*pq)[0 : n+1]
	item := x.(*priorityQueueItem)
	item.Index = n
	(*pq)[n] = item
}

func (pq *priorityQueue) Pop() interface{} {
	n := len(*pq)
	c := cap(*pq)
	if n < (c/2) && c > 25 {
		npq := make(priorityQueue, n, c/2)
		copy(npq, *pq)
		*pq = npq
	}
	item := (*pq)[n-1]
	item.Index = -1
	*pq = (*pq)[0 : n-1]
	return item
}

func (pq *priorityQueue) PeekAndShift(max int64) (*priorityQueueItem, int64) {
	if pq.Len() == 0 {
		return nil, 0
	}

	item := (*pq)[0]
	if item.Priority > max {
		return nil, item.Priority - max
	}
	heap.Remove(pq, 0)

	return item, 0
}

// The end of PriorityQueue implementation.

// timerDelayQueue is an unbounded blocking queue of *Delayed* elements, in which
// an element can only be taken when its delay has expired. The head of the
// queue is the *Delayed* element whose delay expired furthest in the past.
type timerDelayQueue struct {
	C chan interface{}

	mu sync.Mutex
	pq priorityQueue

	// Similar to the sleeping state of runtime.Timers.
	sleeping int32
	wakeupC  chan struct{}
}

// newTimerDelayQueue creates an instance of delayQueue with the specified size.
func newTimerDelayQueue(size int) *timerDelayQueue {
	return &timerDelayQueue{
		C:       make(chan interface{}),
		pq:      newPriorityQueue(size),
		wakeupC: make(chan struct{}),
	}
}

// Offer inserts the element into the current queue.
func (dq *timerDelayQueue) Offer(elem interface{}, expiration int64) {
	item := &priorityQueueItem{Value: elem, Priority: expiration}

	dq.mu.Lock()
	heap.Push(&dq.pq, item)
	index := item.Index
	dq.mu.Unlock()

	if index == 0 {
		// A new item with the earliest expiration is added.
		if atomic.CompareAndSwapInt32(&dq.sleeping, 1, 0) {
			dq.wakeupC <- struct{}{}
		}
	}
}

// Poll starts an infinite loop, in which it continually waits for an element
// to expire and then send the expired element to the channel C.
func (dq *timerDelayQueue) Poll(exitC chan struct{}, nowF func() int64) {
	for {
		now := nowF()

		dq.mu.Lock()
		item, delta := dq.pq.PeekAndShift(now)
		if item == nil {
			// No items left or at least one item is pending.

			// We must ensure the atomicity of the whole operation, which is
			// composed of the above PeekAndShift and the following StoreInt32,
			// to avoid possible race conditions between Offer and Poll.
			atomic.StoreInt32(&dq.sleeping, 1)
		}
		dq.mu.Unlock()

		if item == nil {
			if delta == 0 {
				// No items left.
				select {
				case <-dq.wakeupC:
					// Wait until a new item is added.
					continue
				case <-exitC:
					goto exit
				}
			} else if delta > 0 {
				// At least one item is pending.
				select {
				case <-dq.wakeupC:
					// A new item with an "earlier" expiration than the current "earliest" one is added.
					continue
				case <-time.After(time.Duration(delta) * time.Millisecond):
					// The current "earliest" item expires.

					// Reset the sleeping state since there's no need to receive from wakeupC.
					if atomic.SwapInt32(&dq.sleeping, 0) == 0 {
						// A caller of Offer() is being blocked on sending to wakeupC,
						// drain wakeupC to unblock the caller.
						<-dq.wakeupC
					}
					continue
				case <-exitC:
					goto exit
				}
			}
		}

		select {
		case dq.C <- item.Value:
			// The expired element has been sent out successfully.
		case <-exitC:
			goto exit
		}
	}

exit:
	// Reset the states
	atomic.StoreInt32(&dq.sleeping, 0)
}
