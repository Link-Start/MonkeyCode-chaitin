package apikey

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (p *Postgres) Create(ctx context.Context, key Key, keyHash string) (Key, error) {
	err := p.pool.QueryRow(ctx, `
		INSERT INTO api_keys (user_id, name, key_prefix, key_hash, scopes, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, name, key_prefix, scopes, expires_at,
			last_used_at, created_at, revoked_at
	`, key.UserID, key.Name, key.Prefix, keyHash, key.Scopes, key.ExpiresAt).Scan(
		&key.ID, &key.UserID, &key.Name, &key.Prefix, &key.Scopes,
		&key.ExpiresAt, &key.LastUsedAt, &key.CreatedAt, &key.RevokedAt,
	)
	if err != nil {
		return Key{}, fmt.Errorf("保存调用密钥: %w", err)
	}
	return key, nil
}

func (p *Postgres) ListByUser(ctx context.Context, userID string) ([]Key, error) {
	return p.list(ctx, `
		SELECT k.id, k.user_id, u.name, k.name, k.key_prefix, k.scopes,
			k.expires_at, k.last_used_at, k.created_at, k.revoked_at
		FROM api_keys k
		JOIN users u ON u.id = k.user_id
		WHERE k.user_id = $1
		ORDER BY k.created_at DESC
	`, userID)
}

func (p *Postgres) List(ctx context.Context, userID string) ([]Key, error) {
	if userID != "" {
		return p.ListByUser(ctx, userID)
	}
	return p.list(ctx, `
		SELECT k.id, k.user_id, u.name, k.name, k.key_prefix, k.scopes,
			k.expires_at, k.last_used_at, k.created_at, k.revoked_at
		FROM api_keys k
		JOIN users u ON u.id = k.user_id
		ORDER BY k.created_at DESC
	`)
}

func (p *Postgres) list(ctx context.Context, query string, args ...any) ([]Key, error) {
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询调用密钥: %w", err)
	}
	defer rows.Close()
	keys := make([]Key, 0)
	for rows.Next() {
		var key Key
		if err := rows.Scan(
			&key.ID, &key.UserID, &key.UserName, &key.Name, &key.Prefix,
			&key.Scopes, &key.ExpiresAt, &key.LastUsedAt, &key.CreatedAt, &key.RevokedAt,
		); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (p *Postgres) Revoke(ctx context.Context, id, userID string) error {
	query := `UPDATE api_keys SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`
	args := []any{id}
	if userID != "" {
		query += ` AND user_id = $2`
		args = append(args, userID)
	}
	result, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("撤销调用密钥: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) Authenticate(ctx context.Context, keyHash, scope string) (string, error) {
	var id, userID string
	err := p.pool.QueryRow(ctx, `
		SELECT k.id, k.user_id
		FROM api_keys k
		JOIN users u ON u.id = k.user_id
		WHERE k.key_hash = $1
			AND $2 = ANY(k.scopes)
			AND k.revoked_at IS NULL
			AND k.expires_at > now()
			AND u.status = 'active'
			AND u.deleted_at IS NULL
	`, keyHash, scope).Scan(&id, &userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrInvalidKey
	}
	if err != nil {
		return "", fmt.Errorf("验证调用密钥: %w", err)
	}
	_, _ = p.pool.Exec(ctx, `
		UPDATE api_keys SET last_used_at = now()
		WHERE id = $1
			AND (last_used_at IS NULL OR last_used_at < now() - interval '5 minutes')
	`, id)
	return userID, nil
}
