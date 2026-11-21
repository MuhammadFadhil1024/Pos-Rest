package repository

import (
	"context"
	"database/sql"

	"pos-backend/internal/domain"
	"pos-backend/internal/repository/mysql"
)

type registrationUnitOfWork struct {
    db *sql.DB
}

func NewRegistrationUnitOfWork(
    db *sql.DB,
) domain.RegistrationUnitOfWork {
    return &registrationUnitOfWork{db: db}
}

func (u *registrationUnitOfWork) WithinTransaction(
    ctx context.Context,
    fn func(
        repos domain.RegistrationRepositories,
    ) error,
) error {
    tx, err := u.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    defer tx.Rollback()

    repos := domain.RegistrationRepositories{
        UserRepo:         mysql.NewUserRepository(tx),
        BranchRepo:       mysql.NewBranchRepository(tx),
        UserBranchesRepo: mysql.NewUserBranchesRepository(tx),
    }

    if err := fn(repos); err != nil {
        return err
    }

    return tx.Commit()
}
