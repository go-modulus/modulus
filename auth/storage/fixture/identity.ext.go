package fixture

import (
	"context"
	"testing"
)

// CleanupAllOfAccount calls testing.TB.Cleanup() function with providing a callback inside it.
// This callback will delete all records from the table by the UserID field.
func (f *IdentityFixture) CleanupAllOfAccount(tb testing.TB) *IdentityFixture {
	tb.Cleanup(
		func() {
			query := `DELETE FROM auth.identity WHERE identity.account_id = $1`
			_, err := f.db.Exec(context.Background(), query, f.entity.AccountID)

			if err != nil {
				tb.Fatalf("failed to cleanup Identities of user: %v", err)
			}
		},
	)

	return f
}

// PullUpdatesLastAccountIdentity gets the last Identity of the account and updates the fixture entity.
// This method is useful when you need to get the data from the database after registering the account, when you don't have identity ID
func (f *IdentityFixture) PullUpdatesLastAccountIdentity(tb testing.TB) *IdentityFixture {

	query := `SELECT id FROM auth.identity WHERE identity.account_id = $1 ORDER BY created_at DESC LIMIT 1`
	rows, err := f.db.Query(context.Background(), query, f.entity.AccountID)

	if err != nil {
		tb.Fatalf("failed to get the last identity of the account %s: %v", f.entity.AccountID, err)
	}

	defer rows.Close()

	if !rows.Next() {
		tb.Fatalf("no identity found for the account %s", f.entity.AccountID)
	}

	err = rows.Scan(&f.entity.ID)
	if err != nil {
		tb.Fatalf("failed to scan the last identity of the account %s: %v", f.entity.AccountID, err)
	}

	return f.PullUpdates(tb)
}
