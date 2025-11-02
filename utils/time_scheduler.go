package utils

import (
	"context"
	"sync"
	"time"
)

const ExpectedUnixTimeNotSet int64 = 0

// ElementWithCallback ..
type ElementWithCallback[T any] struct {
	identifier string
	value      *T
	callback   func(value *T) (appendToQueue bool)
}

// TimeScheduler ..
type TimeScheduler[T any] struct {
	mutex   *sync.Mutex
	trans   *Transaction
	mapping map[string]int64

	suggestedSeconds int64
	checkTimeSeconds int64

	headUnixTime   int64
	scheduledQueue []*ElementWithCallback[T]
	pendingQueue   []*ElementWithCallback[T]
}

// NewTimeScheduler ..
func NewTimeScheduler[T any](ctx context.Context, suggestedSeconds int64, checkTimeSeconds int64) *TimeScheduler[T] {
	t := &TimeScheduler[T]{
		mutex:            new(sync.Mutex),
		trans:            NewTransaction(),
		mapping:          make(map[string]int64),
		suggestedSeconds: suggestedSeconds,
		checkTimeSeconds: checkTimeSeconds,
		headUnixTime:     time.Now().Unix(),
		scheduledQueue:   nil,
		pendingQueue:     nil,
	}

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(t.checkTimeSeconds))
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
			t.ticker()
		}
	}()

	return t
}

// append ..
func (t *TimeScheduler[T]) append(element *ElementWithCallback[T], index int64, useLock bool) {
	if useLock {
		t.mutex.Lock()
		defer t.mutex.Unlock()
	}

	for len(t.scheduledQueue) <= int(index) {
		t.scheduledQueue = append(t.scheduledQueue, nil)
	}
	for i := index; i >= 0; i-- {
		if t.scheduledQueue[i] == nil {
			t.scheduledQueue[i] = element
			t.mapping[element.identifier] = t.headUnixTime + i*t.checkTimeSeconds
			return
		}
	}

	t.pendingQueue = append(t.pendingQueue, element)
}

// append ..
func (t *TimeScheduler[T]) appendWithAuto(element *ElementWithCallback[T], useLock bool) {
	if useLock {
		t.mutex.Lock()
		defer t.mutex.Unlock()
	}

	currentTime := time.Now().Unix()
	suggestedTime := currentTime + t.suggestedSeconds
	suggestedIndex := (suggestedTime-t.headUnixTime+t.checkTimeSeconds-1)/t.checkTimeSeconds - 1

	t.append(element, suggestedIndex, false)
}

// appendWithTime ..
func (t *TimeScheduler[T]) appendWithTime(element *ElementWithCallback[T], expectedUnixTime int64, useLock bool) (success bool) {
	if useLock {
		t.mutex.Lock()
		defer t.mutex.Unlock()
	}

	if expectedUnixTime < t.headUnixTime {
		return false
	}

	delta := expectedUnixTime - t.headUnixTime
	ordinal := (delta + t.checkTimeSeconds - 1) / t.checkTimeSeconds
	index := ordinal - 1
	if index < 0 {
		return false
	}

	t.append(element, index, false)
	return true
}

// compact ..
func (t *TimeScheduler[T]) compact(useLock bool) {
	if useLock {
		t.mutex.Lock()
		defer t.mutex.Unlock()
	}

	t.mapping = make(map[string]int64)
	t.headUnixTime = time.Now().Unix()

	for index, element := range t.scheduledQueue {
		if element != nil {
			triggerUnixTime := t.headUnixTime + int64(index)*t.checkTimeSeconds
			t.mapping[element.identifier] = triggerUnixTime
		}
	}
}

// ticker ..
func (t *TimeScheduler[T]) ticker() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	appendFunc := func(e *ElementWithCallback[T]) {
		defer t.trans.Unlock(e.identifier)

		if e.callback(e.value) {
			t.appendWithAuto(e, true)
			return
		}

		t.mutex.Lock()
		delete(t.mapping, e.identifier)
		t.mutex.Unlock()
	}

	if len(t.pendingQueue) > 0 {
		for index, element := range t.pendingQueue {
			t.trans.Lock(element.identifier)
			t.pendingQueue[index] = nil
			go appendFunc(element)
		}
		t.pendingQueue = nil
	}

	if len(t.scheduledQueue) > 0 {
		if element := t.scheduledQueue[0]; element != nil {
			t.trans.Lock(element.identifier)
			t.scheduledQueue[0] = nil
			go appendFunc(element)
		}
	}

	t.headUnixTime += t.checkTimeSeconds
	t.scheduledQueue = append(t.scheduledQueue, nil)
	t.scheduledQueue = t.scheduledQueue[1:]
	if t.headUnixTime != time.Now().Unix() {
		t.compact(false)
	}
}

// Append ..
func (t *TimeScheduler[T]) Append(
	identifier string,
	element *T,
	callback func(value *T) (appendToQueue bool),
	expectedUnixTime int64,
) (alreadyHit bool, success bool) {
	for {
		t.mutex.Lock()
		if !t.trans.TryLock(identifier) {
			t.mutex.Unlock()
			t.trans.Lock(identifier)
			t.trans.Unlock(identifier)
			continue
		}
		break
	}
	defer t.mutex.Unlock()
	defer t.trans.Unlock(identifier)

	if _, ok := t.mapping[identifier]; ok {
		return true, false
	}
	for _, value := range t.pendingQueue {
		if value.identifier == identifier {
			return true, false
		}
	}

	e := &ElementWithCallback[T]{
		identifier: identifier,
		value:      element,
		callback:   callback,
	}
	if expectedUnixTime == ExpectedUnixTimeNotSet {
		t.appendWithAuto(e, false)
		return false, true
	}
	return false, t.appendWithTime(e, expectedUnixTime, false)
}

// Delete ..
func (t *TimeScheduler[T]) Delete(identifier string) {
	for {
		t.mutex.Lock()
		if !t.trans.TryLock(identifier) {
			t.mutex.Unlock()
			t.trans.Lock(identifier)
			t.trans.Unlock(identifier)
			continue
		}
		break
	}
	defer t.mutex.Unlock()
	defer t.trans.Unlock(identifier)

	if len(t.pendingQueue) > 0 {
		pendingQueue := make([]*ElementWithCallback[T], 0)
		for _, value := range t.pendingQueue {
			if value.identifier != identifier {
				pendingQueue = append(pendingQueue, value)
			}
		}
		t.pendingQueue = pendingQueue
	}

	if triggerUnixTime, ok := t.mapping[identifier]; ok {
		idx := (triggerUnixTime - t.headUnixTime) / t.checkTimeSeconds
		t.scheduledQueue[idx] = nil
		delete(t.mapping, identifier)
	}
}
