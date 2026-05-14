package save

import "encoding/json"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) Save(slotID string, g GameAccess) error {
	snap := Build(g)
	data, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	return s.Repo.Save(slotID, data)
}

func (s *Service) Load(slotID string) (*Snapshot, error) {
	data, err := s.Repo.Load(slotID)
	if err != nil {
		return nil, err
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

func (s *Service) HasSave(slotID string) bool {
	return s.Repo.Exists(slotID)
}

func (s *Service) Delete(slotID string) error {
	return s.Repo.Delete(slotID)
}
