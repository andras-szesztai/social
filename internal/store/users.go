package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/andras-szesztai/social/internal/utils"
	"github.com/lib/pq"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

type User struct {
	ID        int64     `json:"id" example:"1"`
	Username  string    `json:"username" example:"john_doe"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}

func (s *UserStore) Create(ctx context.Context, user *User) (*User, error) {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	row := s.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password)
	err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserStore) Read(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) Follow(ctx context.Context, userID, followerID int64) error {
	query := `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
	`

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	return err
}

func (s *UserStore) Unfollow(ctx context.Context, userID, followerID int64) error {
	query := `
		DELETE FROM followers WHERE user_id = $1 AND follower_id = $2
	`

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	return err
}

type UserFeed struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	UserID       int64     `json:"user_id"`
	Username     string    `json:"username"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
	Tags         []string  `json:"tags"`
	CommentCount int64     `json:"comment_count"`
}

func (s *UserStore) ReadFeed(ctx context.Context, userID int64, fq utils.FeedQuery) ([]UserFeed, error) {
	query := `
		SELECT 
			p.id, p.user_id, p.title, p.content, p.created_at, p.tags,
			COUNT(c.id) as comment_count,
			u.username
		FROM posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		LEFT JOIN followers f ON f.user_id = p.user_id AND f.follower_id = $1
		WHERE 
			(p.user_id = $1 OR f.follower_id = $1) AND 
			(p.title ILIKE '%' || $2 || '%' OR p.content ILIKE '%' || $2 || '%') AND
			(p.tags @> $3 OR $3 = '{}')
		GROUP BY p.id, u.username
		ORDER BY p.created_at ` + fq.Sort + ` 
		OFFSET $4 LIMIT $5	
	`

	rows, err := s.db.QueryContext(ctx, query, userID, fq.Search, pq.Array(fq.Tags), fq.Offset, fq.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []UserFeed
	for rows.Next() {
		var item UserFeed
		if err := rows.Scan(&item.ID, &item.UserID, &item.Title, &item.Content, &item.CreatedAt, pq.Array(&item.Tags), &item.CommentCount, &item.Username); err != nil {
			return nil, err
		}
		feed = append(feed, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return feed, nil
}
