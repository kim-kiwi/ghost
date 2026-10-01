package repo

import (
	"errors"
)

type OnMemResource interface {
    SetId(id int64)
}
type OnMem[T OnMemResource] struct {
	NextId int64
	Items map[int64]T
}
func NewOnMem[T OnMemResource]() *OnMem[T] {
	return &OnMem[T]{Items:make(map[int64]T)}
}
func (s *OnMem[T]) Create(t T) (int64, error) {
	s.NextId++
	t.SetId(s.NextId)
	s.Items[s.NextId] = t
	return s.NextId, nil
}
func (s *OnMem[T]) List() ([]T, error) {
	ts := make([]T, 0)
	for _, value := range s.Items {
		ts = append(ts, value)
	}
	return ts, nil
}
func (s *OnMem[T]) Get(id int64) (T, error) {
	t, ok := s.Items[id]
	if !ok {
		var zero T
		return zero, errors.New("Item not found")
	}
	return t, nil
}
func (s *OnMem[T]) Delete(id int64) error {
	_, ok := s.Items[id]
	if !ok {
		return errors.New("Item not found")
	}
	delete(s.Items, id)
	return nil
}
