package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"user-management-service/model"

	"github.com/lib/pq"
)

type Service interface {
	CreateUser(ctx context.Context, user model.User) error
	GetUser(ctx context.Context, userID string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}

type service struct {
	db *sql.DB
}

func (s *service) CreateUser(ctx context.Context, user model.User) error {
	parts := strings.Split(user.Email, "@")
	names := parts[:len(parts)-2]
	name := strings.Join(names, "")
	_, err := s.db.ExecContext(ctx, `INSERT INTO users(email, name, password) values($1, $2, $3)`, user.Email, name, user.Password)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" { // Unique violation error code in PostgreSQL
				fmt.Println("Primary key violation: duplicate entry")
				return errors.New("user already exists")
			}
		}
		return err
	}

	return nil
}

func (s *service) GetUser(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	if err := s.db.QueryRowContext(ctx, `SELECT * FROM users where "ID" = $1`, userID).
		Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	row := s.db.QueryRowContext(ctx, `SELECT * FROM users where email = $1`, email)
	if err := row.Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
		return nil, err
	}

	return &user, nil
}

func NewService(db *sql.DB) Service {
	return &service{db: db}
}
