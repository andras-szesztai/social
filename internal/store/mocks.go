package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/andras-szesztai/social/internal/utils"
)

func NewMockStore() *Store {
	return &Store{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct {
	ReadByIDFunc func(ctx context.Context, id int64) (*User, error)
}

func (m *MockUserStore) Create(ctx context.Context, tx *sql.Tx, user *User) (*User, error) {
	return nil, nil
}

func (m *MockUserStore) ReadByID(ctx context.Context, id int64) (*User, error) {
	return &User{
		ID: id,
	}, nil
}

func (m *MockUserStore) ReadByEmail(ctx context.Context, email string) (*User, error) {
	return nil, nil
}

func (m *MockUserStore) Follow(ctx context.Context, userID, followerID int64) error {
	return nil
}

func (m *MockUserStore) Unfollow(ctx context.Context, userID, followerID int64) error {
	return nil
}

func (m *MockUserStore) ReadFeed(ctx context.Context, userID int64, fq utils.FeedQuery) ([]UserFeed, error) {
	return nil, nil
}

func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExpiry time.Duration) error {
	return nil
}

func (m *MockUserStore) Activate(ctx context.Context, userID int64, token string) error {
	return nil
}

func (m *MockUserStore) Delete(ctx context.Context, id int64) error {
	return nil
}
