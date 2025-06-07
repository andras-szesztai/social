package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/andras-szesztai/social/internal/utils"
)

var (
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrNotFound              = errors.New("not found")
	ErrInvitationExpired     = errors.New("invitation expired")
)

type Store struct {
	Users interface {
		Create(ctx context.Context, tx *sql.Tx, user *User) (*User, error)
		ReadByID(ctx context.Context, id int64) (*User, error)
		ReadByEmail(ctx context.Context, email string) (*User, error)
		Follow(ctx context.Context, userID, followerID int64) error
		Unfollow(ctx context.Context, userID, followerID int64) error
		ReadFeed(ctx context.Context, userID int64, fq utils.FeedQuery) ([]UserFeed, error)
		CreateAndInvite(ctx context.Context, user *User, token string, invitationExpiry time.Duration) error
		Activate(ctx context.Context, userID int64, token string) error
		Delete(ctx context.Context, id int64) error
	}
	Posts interface {
		Create(ctx context.Context, post *Post) (*Post, error)
		Read(ctx context.Context, id int64) (*Post, error)
		Update(ctx context.Context, post *Post) (*Post, error)
		Delete(ctx context.Context, id int64) error
	}
	Comments interface {
		Create(ctx context.Context, comment *Comment) (*Comment, error)
		Read(ctx context.Context, id int64) (*Comment, error)
		ReadByPostID(ctx context.Context, postID int64) ([]Comment, error)
		Update(ctx context.Context, comment *Comment) (*Comment, error)
		Delete(ctx context.Context, id int64) error
	}
	Roles interface {
		ReadByName(ctx context.Context, name string) (*Role, error)
	}
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Users:    NewUserStore(db),
		Posts:    NewPostStore(db),
		Comments: NewCommentStore(db),
		Roles:    NewRoleStore(db),
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("failed to rollback transaction: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
