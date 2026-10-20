package mysql

import (
	"database/sql"
	"pos-backend/internal/domain"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *transactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) GetTransactionByBranchId(branch_id int64) ([]*domain.Transaction, error) {
	return nil, nil
}
