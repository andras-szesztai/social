package store

import (
	"context"
	"database/sql"
)

type Store struct {
	Posts interface {
		Create(ctx context.Context, post *Post) (*Post, error)
		Read(ctx context.Context, id int64) (*Post, error)
		Update(ctx context.Context, post *Post) (*Post, error)
		Delete(ctx context.Context, id int64) error
	}
	Users interface {
		Create(ctx context.Context, user *User) (*User, error)
		Read(ctx context.Context, id int64) (*User, error)
		Follow(ctx context.Context, userID, followerID int64) error
		Unfollow(ctx context.Context, userID, followerID int64) error
		ReadFeed(ctx context.Context, userID int64) ([]UserFeed, error)
	}
	Comments interface {
		Create(ctx context.Context, comment *Comment) (*Comment, error)
		Read(ctx context.Context, id int64) (*Comment, error)
		ReadByPostID(ctx context.Context, postID int64) ([]Comment, error)
		Update(ctx context.Context, comment *Comment) (*Comment, error)
		Delete(ctx context.Context, id int64) error
	}
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Posts:    NewPostStore(db),
		Users:    NewUserStore(db),
		Comments: NewCommentStore(db),
	}
}
