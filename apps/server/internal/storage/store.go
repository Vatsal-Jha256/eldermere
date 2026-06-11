package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
)

type Store interface {
	CreatePlayerSession(ctx context.Context, displayName string) (PlayerSession, error)
	VerifyPlayerSession(ctx context.Context, playerID string, token string) (bool, error)
	LoadPlayerState(ctx context.Context, playerID string) (game.PersistentState, bool, error)
	SavePlayerState(ctx context.Context, playerID string, state game.PersistentState) error
	Close()
}

type PlayerSession struct {
	PlayerID    string `json:"player_id"`
	DisplayName string `json:"display_name"`
	Token       string `json:"token"`
}

type storedSession struct {
	DisplayName string
	TokenHash   string
}

type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]storedSession
	states   map[string]game.PersistentState
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: map[string]storedSession{},
		states:   map[string]game.PersistentState{},
	}
}

func (s *MemoryStore) CreatePlayerSession(_ context.Context, displayName string) (PlayerSession, error) {
	session, err := newPlayerSession(displayName)
	if err != nil {
		return PlayerSession{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.PlayerID] = storedSession{
		DisplayName: session.DisplayName,
		TokenHash:   hashToken(session.Token),
	}
	return session, nil
}

func (s *MemoryStore) VerifyPlayerSession(_ context.Context, playerID string, token string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[playerID]
	if !ok {
		return false, nil
	}
	return session.TokenHash == hashToken(token), nil
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

func newPlayerSession(displayName string) (PlayerSession, error) {
	playerID, err := randomHex(16)
	if err != nil {
		return PlayerSession{}, err
	}
	token, err := randomHex(32)
	if err != nil {
		return PlayerSession{}, err
	}
	if displayName == "" {
		displayName = "Wanderer"
	}

	return PlayerSession{
		PlayerID:    playerID,
		DisplayName: displayName,
		Token:       token,
	}, nil
}

func randomHex(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
