package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/andras-szesztai/social/internal/utils"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

type User struct {
	ID          int64     `json:"id" example:"1"`
	Username    string    `json:"username" example:"john_doe"`
	Email       string    `json:"email" example:"john.doe@example.com"`
	Password    password  `json:"-"`
	CreatedAt   time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
	IsActivated bool      `json:"is_activated" example:"true"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.plaintext = &plaintext
	p.hash = hash

	return nil
}

func (p *password) Compare(plaintext string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(plaintext))
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) (*User, error) {
	query := `	
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	row := tx.QueryRowContext(ctx, query, user.Username, user.Email, user.Password.hash)
	err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		switch err.Error() {
		case "pq: duplicate key value violates unique constraint \"users_email_key\"":
			return nil, ErrEmailAlreadyExists
		case "pq: duplicate key value violates unique constraint \"users_username_key\"":
			return nil, ErrUsernameAlreadyExists
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) ReadByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, query, id)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) ReadByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email,  password, created_at,updated_at
		FROM users
		WHERE email = $1 AND activated = true
	`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx, query, email)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password.hash, &user.CreatedAt, &user.UpdatedAt)
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

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	return err
}

func (s *UserStore) Unfollow(ctx context.Context, userID, followerID int64) error {
	query := `
		DELETE FROM followers WHERE user_id = $1 AND follower_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

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

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

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

func (s *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExpiry time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {

		user, err := s.Create(ctx, tx, user)
		if err != nil {
			return err
		}

		query := `
			INSERT INTO user_invitations (user_id, token, expires_at)	
			VALUES ($1, $2, $3)
		`

		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		_, err = tx.ExecContext(ctx, query, user.ID, token, time.Now().Add(invitationExpiry))
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) Activate(ctx context.Context, userID int64, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		query := `
			SELECT user_id, token, expires_at
			FROM user_invitations
			WHERE token = $1
		`

		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		row := tx.QueryRowContext(ctx, query, token)
		var invitationUserID int64
		var invitationToken string
		var invitationExpiresAt time.Time
		err := row.Scan(&invitationUserID, &invitationToken, &invitationExpiresAt)
		if err != nil {
			return err
		}
		if userID != invitationUserID {
			return ErrNotFound
		}
		if time.Now().After(invitationExpiresAt) {
			return ErrInvitationExpired
		}

		query = `
			UPDATE users SET activated = true WHERE id = $1
		`

		_, err = tx.ExecContext(ctx, query, userID)
		if err != nil {
			return err
		}

		query = `
			DELETE FROM user_invitations WHERE token = $1
		`

		_, err = tx.ExecContext(ctx, query, token)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) Delete(ctx context.Context, id int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		query := `
			DELETE FROM users WHERE id = $1
		`

		_, err := tx.ExecContext(ctx, query, id)
		if err != nil {
			return err
		}

		return nil
	})
}
