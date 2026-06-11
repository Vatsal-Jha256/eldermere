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
		create table if not exists player_accounts (
			player_id text primary key,
			display_name text not null,
			token_hash text not null,
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now()
		);

		create table if not exists player_states (
			player_id text primary key references player_accounts(player_id) on delete cascade,
			room_id text not null,
			party jsonb not null default '[]'::jsonb,
			items jsonb not null default '[]'::jsonb,
			factions jsonb not null default '{}'::jsonb,
			quest_started boolean not null default false,
			quest_completed boolean not null default false,
			quest_variant text not null default '',
			updated_at timestamptz not null default now()
		);
		alter table player_states add column if not exists factions jsonb not null default '{}'::jsonb;
		alter table player_states add column if not exists quest_variant text not null default ''
	`)
	return err
}

func (s *PostgresStore) CreatePlayerSession(ctx context.Context, displayName string) (PlayerSession, error) {
	session, err := newPlayerSession(displayName)
	if err != nil {
		return PlayerSession{}, err
	}

	_, err = s.pool.Exec(ctx, `
		insert into player_accounts (
			player_id,
			display_name,
			token_hash,
			created_at,
			updated_at
		)
		values ($1, $2, $3, $4, $4)
	`, session.PlayerID, session.DisplayName, hashToken(session.Token), nowUTC())
	if err != nil {
		return PlayerSession{}, err
	}

	return session, nil
}

func (s *PostgresStore) VerifyPlayerSession(ctx context.Context, playerID string, token string) (bool, error) {
	var tokenHash string
	err := s.pool.QueryRow(ctx, `
		select token_hash
		from player_accounts
		where player_id = $1
	`, playerID).Scan(&tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return tokenHash == hashToken(token), nil
}

func (s *PostgresStore) PlayerDisplayName(ctx context.Context, playerID string) (string, bool, error) {
	var displayName string
	err := s.pool.QueryRow(ctx, `
		select display_name
		from player_accounts
		where player_id = $1
	`, playerID).Scan(&displayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	return displayName, true, nil
}

func (s *PostgresStore) LoadPlayerState(ctx context.Context, playerID string) (game.PersistentState, bool, error) {
	var state game.PersistentState
	var partyJSON []byte
	var itemsJSON []byte
	var factionsJSON []byte

	err := s.pool.QueryRow(ctx, `
		select room_id, party, items, factions, quest_started, quest_completed, quest_variant
		from player_states
		where player_id = $1
	`, playerID).Scan(&state.RoomID, &partyJSON, &itemsJSON, &factionsJSON, &state.Quest.Started, &state.Quest.Completed, &state.Quest.Variant)
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
	if err := json.Unmarshal(factionsJSON, &state.Factions); err != nil {
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
	factionsJSON, err := json.Marshal(state.Factions)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		insert into player_states (
			player_id,
			room_id,
			party,
			items,
			factions,
			quest_started,
			quest_completed,
			quest_variant,
			updated_at
		)
		values ($1, $2, $3::jsonb, $4::jsonb, $5::jsonb, $6, $7, $8, now())
		on conflict (player_id) do update set
			room_id = excluded.room_id,
			party = excluded.party,
			items = excluded.items,
			factions = excluded.factions,
			quest_started = excluded.quest_started,
			quest_completed = excluded.quest_completed,
			quest_variant = excluded.quest_variant,
			updated_at = now()
	`, playerID, state.RoomID, string(partyJSON), string(itemsJSON), string(factionsJSON), state.Quest.Started, state.Quest.Completed, state.Quest.Variant)
	return err
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}
