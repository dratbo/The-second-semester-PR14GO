package store

type ProcessedStore struct {
	items map[string]bool
}

func NewProcessedStore() *ProcessedStore {
	return &ProcessedStore{
		items: make(map[string]bool),
	}
}

func (s *ProcessedStore) Exists(id string) bool {
	return s.items[id]
}

func (s *ProcessedStore) MarkDone(id string) {
	s.items[id] = true
}
