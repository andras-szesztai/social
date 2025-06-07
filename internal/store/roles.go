package store

import (
	"context"
	"database/sql"
)

type RoleStore struct {
	db *sql.DB
}

type Role struct {
	ID          int64  `json:"id" example:"1"`
	Name        string `json:"name" example:"user"`
	Level       int64  `json:"level" example:"1"`
	Description string `json:"description" example:"User role"`
}

func NewRoleStore(db *sql.DB) *RoleStore {
	return &RoleStore{db: db}
}

func (s *RoleStore) ReadByName(ctx context.Context, name string) (*Role, error) {
	query := `SELECT id, name, level, description FROM roles WHERE name = $1`
	row := s.db.QueryRowContext(ctx, query, name)

	var role Role
	err := row.Scan(&role.ID, &role.Name, &role.Level, &role.Description)
	if err != nil {
		return nil, err
	}

	return &role, nil
}
