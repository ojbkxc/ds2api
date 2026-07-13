package localtool

import (
	"sync"
)

type MemoryType string

const (
	MemoryTypeUser      MemoryType = "user"
	MemoryTypeFeedback  MemoryType = "feedback"
	MemoryTypeTopic     MemoryType = "topic"
	MemoryTypeReference MemoryType = "reference"
)

type Memory struct {
	ID          int        `json:"id"`
	Type        MemoryType `json:"type"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Content     string     `json:"content"`
	Tags        []string   `json:"tags"`
	Pinned      bool       `json:"pinned"`
	CreatedAt   int64      `json:"created_at"`
	UpdatedAt   int64      `json:"updated_at"`
}

type NewMemory struct {
	Type        MemoryType `json:"type"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Content     string     `json:"content"`
	Tags        []string   `json:"tags"`
	Pinned      bool       `json:"pinned"`
}

type MemoryStorage interface {
	SaveMemory(memory NewMemory) (*Memory, error)
	GetMemoryByID(id int) (*Memory, error)
	UpdateMemory(memory Memory) error
	DeleteMemory(id int) error
	ListMemories() ([]Memory, error)
	ListMemoriesByType(memoryType MemoryType) ([]Memory, error)
}

type InMemoryStorage struct {
	memories []Memory
	nextID   int
	mu       sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		memories: make([]Memory, 0),
		nextID:   1,
	}
}

func (s *InMemoryStorage) SaveMemory(memory NewMemory) (*Memory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := timeNowUnix()
	m := Memory{
		ID:          s.nextID,
		Type:        memory.Type,
		Name:        memory.Name,
		Description: memory.Description,
		Content:     memory.Content,
		Tags:        memory.Tags,
		Pinned:      memory.Pinned,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.nextID++
	s.memories = append(s.memories, m)
	return &m, nil
}

func (s *InMemoryStorage) GetMemoryByID(id int) (*Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, m := range s.memories {
		if m.ID == id {
			return &m, nil
		}
	}
	return nil, &MemoryNotFoundError{ID: id}
}

func (s *InMemoryStorage) UpdateMemory(memory Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, m := range s.memories {
		if m.ID == memory.ID {
			memory.UpdatedAt = timeNowUnix()
			s.memories[i] = memory
			return nil
		}
	}
	return &MemoryNotFoundError{ID: memory.ID}
}

func (s *InMemoryStorage) DeleteMemory(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, m := range s.memories {
		if m.ID == id {
			s.memories = append(s.memories[:i], s.memories[i+1:]...)
			return nil
		}
	}
	return &MemoryNotFoundError{ID: id}
}

func (s *InMemoryStorage) ListMemories() ([]Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Memory, len(s.memories))
	copy(result, s.memories)
	return result, nil
}

func (s *InMemoryStorage) ListMemoriesByType(memoryType MemoryType) ([]Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Memory, 0)
	for _, m := range s.memories {
		if m.Type == memoryType {
			result = append(result, m)
		}
	}
	return result, nil
}

type MemoryNotFoundError struct{ ID int }

func (e *MemoryNotFoundError) Error() string {
	return "memory not found: " + fmtInt(e.ID)
}

func timeNowUnix() int64 {
	return time.Now().Unix()
}
