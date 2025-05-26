package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type PostStore struct {
	db *sql.DB
}

func NewPostStore(db *sql.DB) *PostStore {
	return &PostStore{db: db}
}

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *PostStore) Create(ctx context.Context, post *Post) (*Post, error) {
	query := `
		INSERT INTO posts (title, content, user_id, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	row := s.db.QueryRowContext(ctx, query, post.Title, post.Content, post.UserID, pq.Array(post.Tags))

	err := row.Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostStore) Get(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT id, title, content, user_id, tags, created_at, updated_at
		FROM posts
		WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)

	var post Post
	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.UserID, pq.Array(&post.Tags), &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &post, nil
}
