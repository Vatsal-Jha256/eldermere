package storage

import (
	"context"
	"sync"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
)

type Store interface {
	LoadPlayerState(ctx context.Context, playerID string) (game.PersistentState, bool, error)
	SavePlayerState(ctx context.Context, playerID string, state game.PersistentState) error
	Close()
}

type MemoryStore struct {
	mu     sync.RWMutex
	states map[string]game.PersistentState
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		states: map[string]game.PersistentState{},
	}
}

func (s *MemoryStore) LoadPlayerState(_ context.Context, playerID string) (game.PersistentState, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.states[playerID]
	return state, ok, nil
}

func (s *MemoryStore) SavePlayerState(_ context.Context, playerID string, state game.PersistentState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.states[playerID] = state
	return nil
}

func (s *MemoryStore) Close() {}
