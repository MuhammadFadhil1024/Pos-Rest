package service

import "pos-backend/internal/domain"

type transactionService struct {
	transactionRepo domain.TransactionRepository
}

func newTransactionService(repo domain.TransactionRepository) domain.TransactionService {
	return &transactionService{
		transactionRepo: repo,
	}
}

func (s *transactionService) GetBranchTransaction(branch_id int64) ([]*domain.Transaction, error) {
	transactions, err := s.transactionRepo.GetTransactionByBranchId(branch_id)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}