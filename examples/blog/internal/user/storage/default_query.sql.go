// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: default_query.sql

package storage

import (
	"context"
)

const ping = `-- name: Ping :many
SELECT 'pong' AS ping`

func (q *Queries) Ping(ctx context.Context) ([]string, error) {
	rows, err := q.db.Query(ctx, ping)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var ping string
		if err := rows.Scan(&ping); err != nil {
			return nil, err
		}
		items = append(items, ping)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
