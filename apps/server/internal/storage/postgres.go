package storage

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/game"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	store := &PostgresStore{pool: pool}
	if err := store.Migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *PostgresStore) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		create table if not exists player_states (
			player_id text primary key,
			room_id text not null,
			party jsonb not null default '[]'::jsonb,
			items jsonb not null default '[]'::jsonb,
			quest_started boolean not null default false,
			quest_completed boolean not null default false,
			updated_at timestamptz not null default now()
		)
	`)
	return err
}

func (s *PostgresStore) LoadPlayerState(ctx context.Context, playerID string) (game.PersistentState, bool, error) {
	var state game.PersistentState
	var partyJSON []byte
	var itemsJSON []byte

	err := s.pool.QueryRow(ctx, `
		select room_id, party, items, quest_started, quest_completed
		from player_states
		where player_id = $1
	`, playerID).Scan(&state.RoomID, &partyJSON, &itemsJSON, &state.Quest.Started, &state.Quest.Completed)
	if errors.Is(err, pgx.ErrNoRows) {
		return game.PersistentState{}, false, nil
	}
	if err != nil {
		return game.PersistentState{}, false, err
	}

	if err := json.Unmarshal(partyJSON, &state.Party); err != nil {
		return game.PersistentState{}, false, err
	}
	if err := json.Unmarshal(itemsJSON, &state.Items); err != nil {
		return game.PersistentState{}, false, err
	}

	return state, true, nil
}

func (s *PostgresStore) SavePlayerState(ctx context.Context, playerID string, state game.PersistentState) error {
	partyJSON, err := json.Marshal(state.Party)
	if err != nil {
		return err
	}
	itemsJSON, err := json.Marshal(state.Items)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		insert into player_states (
			player_id,
			room_id,
			party,
			items,
			quest_started,
			quest_completed,
			updated_at
		)
		values ($1, $2, $3::jsonb, $4::jsonb, $5, $6, now())
		on conflict (player_id) do update set
			room_id = excluded.room_id,
			party = excluded.party,
			items = excluded.items,
			quest_started = excluded.quest_started,
			quest_completed = excluded.quest_completed,
			updated_at = now()
	`, playerID, state.RoomID, string(partyJSON), string(itemsJSON), state.Quest.Started, state.Quest.Completed)
	return err
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}
