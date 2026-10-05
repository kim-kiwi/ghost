package repo

import (
	"errors"
	"sync"
)

type OnMemResource interface {
	SetId(id int64)
}
type OnMem[T OnMemResource] struct {
	nextId int64
	items  map[int64]T
	mu     sync.RWMutex
}

func NewOnMem[T OnMemResource]() *OnMem[T] {
	return &OnMem[T]{items: make(map[int64]T)}
}
func (s *OnMem[T]) Create(t T) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextId++
	t.SetId(s.nextId)
	s.items[s.nextId] = t
	return s.nextId, nil
}
func (s *OnMem[T]) List() ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ts := make([]T, 0)
	for _, value := range s.items {
		ts = append(ts, value)
	}
	return ts, nil
}
func (s *OnMem[T]) Find(predicate func(T) bool) (T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, value := range s.items {
		if predicate(value) {
			return value, nil
		}
	}

	var zero T
	return zero, errors.New("item not found")
}
func (s *OnMem[T]) FindMany(predicate func(T) bool) ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []T
	for _, value := range s.items {
		if predicate(value) {
			result = append(result, value)
		}
	}

	return result, nil
}
func (s *OnMem[T]) Get(id int64) (T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.items[id]
	if !ok {
		var zero T
		return zero, errors.New("item not found")
	}
	return t, nil
}
func (s *OnMem[T]) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.items[id]
	if !ok {
		return errors.New("item not found")
	}
	delete(s.items, id)
	return nil
}
