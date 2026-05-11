package entity

type Store[T any] struct {
	data map[ID]*T
}

func NewStore[T any]() *Store[T] {
	return &Store[T]{data: make(map[ID]*T)}
}

func (s *Store[T]) Set(id ID, c *T) {
	s.data[id] = c
}

func (s *Store[T]) Get(id ID) (*T, bool) {
	c, ok := s.data[id]
	return c, ok
}

func (s *Store[T]) Remove(id ID) {
	delete(s.data, id)
}

func (s *Store[T]) Has(id ID) bool {
	_, ok := s.data[id]
	return ok
}

func (s *Store[T]) Each(fn func(ID, *T)) {
	for id, c := range s.data {
		fn(id, c)
	}
}

func (s *Store[T]) Count() int {
	return len(s.data)
}
