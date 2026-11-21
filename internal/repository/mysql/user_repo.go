package mysql

import (
	"context"
	"database/sql"
	"errors"
	"pos-backend/internal/domain"
)

// PERBAIKAN 1: Huruf u kecil pada userRepository agar cocok dengan fungsi di bawahnya
type userRepository struct {
	db DB
}

// NewUserRepository membuat instance baru untuk repo user MySQL
func NewUserRepository(db DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (email, password, phone, role) VALUES (?, ?, ?, ?)`

	res, err := r.db.ExecContext(ctx, query, user.Email, user.Password, user.PhoneNumber, user.Role)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return r.db.QueryRowContext(ctx, "SELECT id, email, password, COALESCE(phone, ''), role FROM users WHERE id = ?", user.ID).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.PhoneNumber,
		&user.Role,
	)
}

// PERBAIKAN 3: Sesuaikan nama fungsi dengan interface (GetByEmail, bukan GetEmailById)
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {

	query := `SELECT id, email, password, COALESCE(phone, ''), role FROM users WHERE email = ?`

	user := &domain.User{}

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.PhoneNumber,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return user, nil
}