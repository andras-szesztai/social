package store

import (
	"context"
	"database/sql"
)

type Store struct {
	Posts interface {
		Create(ctx context.Context, post *Post) (*Post, error)
		Get(ctx context.Context, id int64) (*Post, error)
	}
	Users interface {
		Create(ctx context.Context, user *User) (*User, error)
	}
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Posts: NewPostStore(db),
		Users: NewUserStore(db),
	}
}
