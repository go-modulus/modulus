package fixture

import (
	"context"
	"testing"
)

// CleanupAllOfIdentity calls testing.TB.Cleanup() function with providing a callback inside it.
// This callback will delete all records from the table by the IdentityID field.
func (f *CredentialFixture) CleanupAllOfIdentity(tb testing.TB) *CredentialFixture {
	tb.Cleanup(
		func() {
			query := `DELETE FROM auth.credential WHERE credential.identity_id = $1`
			_, err := f.db.Exec(context.Background(), query, f.entity.IdentityID)

			if err != nil {
				tb.Fatalf("failed to cleanup Credentials of identity: %v", err)
			}
		},
	)

	return f
}
