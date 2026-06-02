// This file verifies local scheduled-job orchestration decisions for legacy
// jobs while keeping external LDAP and database state out of the unit test.

package jobs

import (
	"context"
	"errors"
	"testing"

	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

func TestMovableChangeContainerAccountIDsOnlyReturnsLDAPSuccesses(t *testing.T) {
	ctx := context.Background()
	target := &entity.Container{Id: 5, Name: legacyContainerAlumni}
	accounts := []*entity.Account{
		{Id: 101, Number: "ok-101"},
		{Id: 102, Number: "bad-102"},
		{Id: 103, Number: "ok-103"},
	}

	ids, failed := movableChangeContainerAccountIDs(ctx, accounts, target, func(_ context.Context, account *entity.Account, _ *entity.Container) error {
		if account.Number == "bad-102" {
			return errors.New("ldap move failed")
		}
		return nil
	})
	if failed != 1 {
		t.Fatalf("expected one LDAP move failure, got %d", failed)
	}
	if len(ids) != 2 || ids[0] != 101 || ids[1] != 103 {
		t.Fatalf("expected only LDAP-success accounts to be locally updateable, got %#v", ids)
	}
}

func TestMovableChangeContainerAccountIDsTreatsMissingAccountAsFailure(t *testing.T) {
	ids, failed := movableChangeContainerAccountIDs(context.Background(), []*entity.Account{nil}, &entity.Container{Name: legacyContainerAlumni}, func(context.Context, *entity.Account, *entity.Container) error {
		t.Fatal("mover should not be called for missing account")
		return nil
	})
	if failed != 1 {
		t.Fatalf("expected missing account to count as failure, got %d", failed)
	}
	if len(ids) != 0 {
		t.Fatalf("expected no local DB update ids for missing account, got %#v", ids)
	}
}
