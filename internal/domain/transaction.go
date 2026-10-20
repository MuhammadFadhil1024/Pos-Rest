package domain

type Transaction struct {
	Services []*Service `json:"service"`
	Tickets  []*Ticket  `json:"ticket"`
}

type TransactionRepository interface {
	GetTransactionByBranchId(branch_id int64) ([]*Transaction, error)
}

type TransactionService interface {
	GetBranchTransaction(branch_id int64) ([]*Transaction, error)
}