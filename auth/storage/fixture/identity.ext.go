package fixture

import (
	"context"
	"testing"
)

// CleanupAllOfUser calls testing.TB.Cleanup() function with providing a callback inside it.
// This callback will delete all records from the table by the UserID field.
func (f *IdentityFixture) CleanupAllOfUser(tb testing.TB) *IdentityFixture {
	tb.Cleanup(
		func() {
			query := `DELETE FROM auth.identity WHERE identity.user_id = $1`
			_, err := f.db.Exec(context.Background(), query, f.entity.UserID)

			if err != nil {
				tb.Fatalf("failed to cleanup Identities of user: %v", err)
			}
		},
	)

	return f
}
