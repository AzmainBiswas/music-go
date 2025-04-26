package server

import (
	"errors"
	"sync"
)

const (
	MaxStackLen int = 100
)

// stack for previous button
type Stack struct {
	lock  sync.Mutex
	array []int64
}

func NewStack() *Stack {
	return &Stack{sync.Mutex{}, make([]int64, 0)}
}

func (s *Stack) Push(v int64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.array)
	if l > MaxStackLen {
		s.array = s.array[10:]
	}

	s.array = append(s.array, v)
}

var ErrEmptyStack error = errors.New("Stack is empty")

func (s *Stack) Pop() (int64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.array)
	if l == 0 {
		return 0, ErrEmptyStack
	}

	res := s.array[l-1]
	s.array = s.array[:l-1]
	return res, nil
}

// Queue for song queue
type Queue struct {
	lock  sync.Mutex
	array []int64
}

func NewQueue() *Queue {
	return &Queue{sync.Mutex{}, make([]int64, 0)}
}

func (q *Queue) Clear() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.array = q.array[:0]
}

func (q *Queue) Enqueue(values ...int64) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.array = append(q.array, values...)
}

var ErrEmptyQueue = errors.New("Queue is empty")

func (q *Queue) Dequeue() (int64, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	l := len(q.array)
	if l == 0 {
		return 0, ErrEmptyQueue
	}

	res := q.array[0]
	q.array = q.array[1:]
	return res, nil
}
