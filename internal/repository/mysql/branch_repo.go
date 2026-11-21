package mysql

import (
	"context"
	"database/sql"
	"errors"
	"pos-backend/internal/domain"
)

type branchRepository struct {
	db DB
}

func NewBranchRepository(db DB) domain.BranchRepository {
	return &branchRepository{db: db}
}

func (r *branchRepository) GetByUserId(ctx context.Context, id int64) ([]*domain.Branch, error) {
	query := `SELECT b.id as branch_id, b.name, b.slug
		FROM branch b
		JOIN mm_user_branch ub ON b.id = ub.branch_id
		WHERE ub.user_id = ?
	`

	rows, err := r.db.QueryContext(ctx, query, id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var branches []*domain.Branch
	for rows.Next() {
		var b domain.Branch
		if err := rows.Scan(&b.ID, &b.Name, &b.Slug); err != nil {
			return nil, err
		}
		branches = append(branches, &b)
	}

	return branches, nil
}

func (r *branchRepository) GetByBranchId(ctx context.Context, branchId int64) (*domain.Branch, error) {
	query := `SELECT b.id, b.name, b.slug
		FROM branch b WHERE b.id = ?
	`
	branch := &domain.Branch{}

	err := r.db.QueryRowContext(ctx, query, branchId).Scan(
		&branch.ID,
		&branch.Name,
		&branch.Slug,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Branch Not found!")
		}
		return nil, err
	}

	return branch, nil
}

func (r *branchRepository) CreateBranch(ctx context.Context, branch *domain.Branch) error {
	query := `INSERT INTO branch (name, slug, company_name, address) VALUES (?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, branch.Name, branch.Slug, branch.CompanyName, branch.Address)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	branch.ID = id
	return r.db.QueryRowContext(ctx, "SELECT id, name, slug, company_name, address FROM branch WHERE id = ?", branch.ID).Scan(
		&branch.ID,
		&branch.Name,
		&branch.Slug,
		&branch.CompanyName,
		&branch.Address,
	)
}
