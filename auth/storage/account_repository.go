package storage

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/auth/repository"
	"github.com/go-modulus/modulus/errors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DefaultAccountRepository struct {
	queries *Queries
	db      *pgxpool.Pool
}

func NewDefaultAccountRepository(db *pgxpool.Pool) repository.AccountRepository {
	return &DefaultAccountRepository{
		queries: New(db),
		db:      db,
	}
}

func (r *DefaultAccountRepository) Create(ctx context.Context, ID uuid.UUID) (repository.Account, error) {
	_, err := r.Get(ctx, ID)
	if err == nil {
		return repository.Account{}, repository.ErrAccountExists
	} else if !errors.Is(err, repository.ErrAccountNotFound) {
		return repository.Account{}, errtrace.Wrap(err)
	}

	storedAccount, err := r.queries.RegisterAccount(
		ctx, ID,
	)

	if err != nil {
		return repository.Account{}, errtrace.Wrap(errors.WithCause(repository.ErrCannotCreateAccount, err))
	}

	return r.Transform(storedAccount), nil
}

func (r *DefaultAccountRepository) Get(ctx context.Context, ID uuid.UUID) (repository.Account, error) {
	res, err := r.queries.FindAccount(ctx, ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Account{}, repository.ErrAccountNotFound
		}
		return repository.Account{}, errtrace.Wrap(err)
	}
	return r.Transform(res), nil
}

func (r *DefaultAccountRepository) RemoveAccount(ctx context.Context, ID uuid.UUID) error {
	_, err := r.Get(ctx, ID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return errtrace.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := New(tx)
	err = qtx.RemoveAccount(ctx, ID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	err = qtx.RemoveCredentialsOfAccount(ctx, ID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	err = qtx.RemoveIdentitiesOfAccount(ctx, ID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return tx.Commit(ctx)
}

func (r *DefaultAccountRepository) BlockAccount(ctx context.Context, ID uuid.UUID) error {
	_, err := r.Get(ctx, ID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return errtrace.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := New(tx)
	err = qtx.BlockAccount(ctx, ID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	err = qtx.BlockIdentitiesOfAccount(ctx, ID)
	if err != nil {
		return errtrace.Wrap(err)
	}
	return tx.Commit(ctx)
}

func (r *DefaultAccountRepository) Transform(
	account Account,
) repository.Account {
	return repository.Account{
		ID:     account.ID,
		Roles:  account.Roles,
		Status: repository.AccountStatus(account.Status),
	}
}

func (r *DefaultAccountRepository) AddRoles(
	ctx context.Context,
	accountID uuid.UUID,
	roles ...string,
) error {
	return r.queries.AddRoles(
		ctx, AddRolesParams{
			ID:    accountID,
			Roles: roles,
		},
	)
}

func (r *DefaultAccountRepository) RemoveRoles(
	ctx context.Context,
	accountID uuid.UUID,
	roles ...string,
) error {
	return r.queries.RemoveRoles(
		ctx, RemoveRolesParams{
			ID:    accountID,
			Roles: roles,
		},
	)
}
