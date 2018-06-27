package server

import (
	"context"
	"database/sql"

	"go.opencensus.io/trace"
)

var _ Repository = &SQLite{}

// SQLite is a SQLite database backed implementation of Repository
type SQLite struct {
	db *sql.DB
}

// NewSQLiteRepository returns a new Repository backed by SQLite.
func NewSQLiteRepository(db *sql.DB) *SQLite {
	return &SQLite{db: db}
}

// Prime our database with fake data for the purpose of this demo.
func (r *SQLite) Prime(ctx context.Context) (err error) {
	ctx, span := trace.StartSpan(ctx, "prime database")

	defer func() {
		if err != nil {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeInternal,
				Message: err.Error(),
			})
		} else {
			span.SetStatus(trace.Status{
				Code:    trace.StatusCodeOK,
				Message: "OK",
			})
		}
		span.End()
	}()

	if _, err = r.db.ExecContext(
		ctx, "CREATE TABLE item (user_id int, name text);",
	); err != nil {
		return
	}

	var tx *sql.Tx
	if tx, err = r.db.BeginTx(ctx, &sql.TxOptions{}); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	for _, item := range []struct {
		userID int
		name   string
	}{
		{1, "item 1"},
		{1, "item 2"},
		{1, "item 3"},
		{1, "item 4"},
		{2, "item A"},
		{2, "item B"},
		{2, "item C"},
		{3, "item I"},
		{3, "item II"},
	} {
		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO item (user_id, name) VALUES (?, ?)`,
			item.userID, item.name,
		); err != nil {
			return
		}
	}

	return
}

// ListItems implements Repository
func (r *SQLite) ListItems(ctx context.Context, userID int64) ([]*Item, error) {
	rows, err := r.db.QueryContext(
		ctx, "SELECT user_id, name FROM item WHERE user_id = ? ORDER BY name;", userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		if err = rows.Scan(&item.UserID, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}
