package save

type Repository interface {
	Save(slotID string, data []byte) error
	Load(slotID string) ([]byte, error)
	Delete(slotID string) error
	Exists(slotID string) bool
}

type MemoryRepository struct {
	data map[string][]byte
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{data: make(map[string][]byte)}
}

func (r *MemoryRepository) Save(slotID string, data []byte) error {
	r.data[slotID] = append([]byte(nil), data...)
	return nil
}

func (r *MemoryRepository) Load(slotID string) ([]byte, error) {
	d, ok := r.data[slotID]
	if !ok {
		return nil, ErrNotFound
	}
	return d, nil
}

func (r *MemoryRepository) Delete(slotID string) error {
	delete(r.data, slotID)
	return nil
}

func (r *MemoryRepository) Exists(slotID string) bool {
	_, ok := r.data[slotID]
	return ok
}
