package server

import (
	"context"
)

// Repository interface
type Repository interface {
	ListItems(ctx context.Context, ownerID int64) ([]*Item, error)
}
