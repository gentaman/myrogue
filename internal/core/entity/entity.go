package entity

type ID int64

const InvalidID ID = 0

type Manager struct {
	next  ID
	alive map[ID]bool
}

func NewManager() *Manager {
	return &Manager{
		next:  2, // 1 is reserved for player
		alive: make(map[ID]bool),
	}
}

func (m *Manager) Create() ID {
	id := m.next
	m.next++
	m.alive[id] = true
	return id
}

func (m *Manager) CreateWithID(id ID) {
	m.alive[id] = true
	if id >= m.next {
		m.next = id + 1
	}
}

func (m *Manager) Destroy(id ID) {
	delete(m.alive, id)
}

func (m *Manager) IsAlive(id ID) bool {
	return m.alive[id]
}

func (m *Manager) All() []ID {
	ids := make([]ID, 0, len(m.alive))
	for id := range m.alive {
		ids = append(ids, id)
	}
	return ids
}

func (m *Manager) Count() int {
	return len(m.alive)
}
