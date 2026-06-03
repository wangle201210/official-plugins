// grass_type.go defines the grass ledger transaction-type enum as a Go named
// type with constants. Each value is persisted in the grass_txn txn_type column
// and identifies the business event that produced the grass change.

package grass

// TxnType is the grass ledger transaction-type enum. Its string value is
// persisted in the grass_txn txn_type column.
type TxnType string

const (
	// TxnTypeCheckin marks a credit from the daily check-in grant.
	TxnTypeCheckin TxnType = "checkin"
	// TxnTypeFeed marks a debit from feeding grass to a cattle.
	TxnTypeFeed TxnType = "feed"
	// TxnTypeStealGain marks a credit from stealing grass from another player.
	TxnTypeStealGain TxnType = "steal_gain"
	// TxnTypeStolenLoss marks a debit from being stolen from by another player.
	TxnTypeStolenLoss TxnType = "stolen_loss"
	// TxnTypeGiftOut marks a debit from gifting grass to another player.
	TxnTypeGiftOut TxnType = "gift_out"
	// TxnTypeGiftIn marks a credit from receiving gifted grass.
	TxnTypeGiftIn TxnType = "gift_in"
)

// String returns the persisted string form of the transaction type.
func (t TxnType) String() string {
	return string(t)
}
