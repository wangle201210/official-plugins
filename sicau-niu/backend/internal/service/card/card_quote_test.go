// card_quote_test.go verifies quote CRUD, the enabled-flag default, content
// validation and paged listing against a gated PostgreSQL database. Database-
// backed assertions are skipped unless LINA_TEST_PGSQL_LINK is set.

package card

import (
	"context"
	"fmt"
	"testing"
)

// newQuoteServiceForTest builds a card service for quote operations. Quote
// operations do not touch the cattle dependency, so a real cattle service is
// still injected to construct the service exactly as production does.
func newQuoteServiceForTest() Service {
	svc, _ := newCardServiceForTest()
	return svc
}

// TestCreateQuoteRejectsBlankContent verifies a blank content is rejected with
// CodeQuoteContentRequired before any write.
func TestCreateQuoteRejectsBlankContent(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCardDB(t, ctx)

	svc := newQuoteServiceForTest()
	_, err := svc.CreateQuote(ctx, &QuoteMutateInput{Content: "   "})
	assertBizCode(t, err, CodeQuoteContentRequired.RuntimeCode())
}

// TestCreateQuoteDefaultsEnabled verifies a quote created without an explicit
// enabled flag persists enabled=1 so it participates in random playback.
func TestCreateQuoteDefaultsEnabled(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCardDB(t, ctx)

	svc := newQuoteServiceForTest()
	id, err := svc.CreateQuote(ctx, &QuoteMutateInput{Content: "Knowledge is power."})
	if err != nil {
		t.Fatalf("create quote failed: %v", err)
	}

	out, err := svc.ListQuote(ctx, &ListQuoteInput{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list quotes failed: %v", err)
	}
	var found *QuoteItem
	for _, item := range out.List {
		if item.Id == id {
			found = item
			break
		}
	}
	if found == nil {
		t.Fatalf("expected created quote %d in the list", id)
	}
	if found.Enabled != 1 {
		t.Fatalf("expected default enabled=1, got %d", found.Enabled)
	}
}

// TestUpdateQuoteChangesContentAndFlag verifies update replaces content and
// applies an explicit enabled flag, while a missing quote returns
// CodeQuoteNotFound.
func TestUpdateQuoteChangesContentAndFlag(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCardDB(t, ctx)

	svc := newQuoteServiceForTest()
	id, err := svc.CreateQuote(ctx, &QuoteMutateInput{Content: "Original"})
	if err != nil {
		t.Fatalf("create quote failed: %v", err)
	}

	disabled := 0
	if err = svc.UpdateQuote(ctx, id, &QuoteMutateInput{Content: "Updated", Enabled: &disabled}); err != nil {
		t.Fatalf("update quote failed: %v", err)
	}

	out, err := svc.ListQuote(ctx, &ListQuoteInput{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list quotes failed: %v", err)
	}
	var found *QuoteItem
	for _, item := range out.List {
		if item.Id == id {
			found = item
			break
		}
	}
	if found == nil {
		t.Fatalf("expected quote %d in the list", id)
	}
	if found.Content != "Updated" {
		t.Fatalf("expected updated content, got %q", found.Content)
	}
	if found.Enabled != 0 {
		t.Fatalf("expected enabled=0 after update, got %d", found.Enabled)
	}

	err = svc.UpdateQuote(ctx, 778899, &QuoteMutateInput{Content: "Nope"})
	assertBizCode(t, err, CodeQuoteNotFound.RuntimeCode())
}

// TestListQuotePaginates verifies the quote list paginates in ID-descending order
// and reports the true total across pages.
func TestListQuotePaginates(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCardDB(t, ctx)

	svc := newQuoteServiceForTest()
	const total = 5
	for i := 0; i < total; i++ {
		content := fmt.Sprintf("Quote %d", i)
		if _, err := svc.CreateQuote(ctx, &QuoteMutateInput{Content: content}); err != nil {
			t.Fatalf("create quote %d failed: %v", i, err)
		}
	}

	page1, err := svc.ListQuote(ctx, &ListQuoteInput{PageNum: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("list page 1 failed: %v", err)
	}
	if page1.Total != total {
		t.Fatalf("expected total %d, got %d", total, page1.Total)
	}
	if len(page1.List) != 2 {
		t.Fatalf("expected 2 rows on page 1, got %d", len(page1.List))
	}
	if page1.List[0].Id <= page1.List[1].Id {
		t.Fatalf("expected ID-descending order, got %d then %d", page1.List[0].Id, page1.List[1].Id)
	}

	page3, err := svc.ListQuote(ctx, &ListQuoteInput{PageNum: 3, PageSize: 2})
	if err != nil {
		t.Fatalf("list page 3 failed: %v", err)
	}
	if len(page3.List) != 1 {
		t.Fatalf("expected 1 row on page 3, got %d", len(page3.List))
	}
}
