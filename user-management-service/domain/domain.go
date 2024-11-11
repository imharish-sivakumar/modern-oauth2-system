package domain

import (
	"context"
	"database/sql"

	"user-management-service/models"
)

type Service interface {
	CreateUser(ctx context.Context, user models.User) error
	GetUser(ctx context.Context, userID string) (*models.User, error)
}

type service struct {
	db *sql.DB
}

func (s *service) CreateUser(ctx context.Context, user models.User) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO users(email, password) values($1, $2)`, user.Email, user.Password)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	if err := s.db.QueryRowContext(ctx, `SELECT * FROM users where id = $1`, userID).
		Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
		return nil, err
	}

	return &user, nil
}

func NewService(db *sql.DB) Service {
	return &service{db: db}
}
