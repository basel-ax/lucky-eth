package entity

import (
	"time"

	"gorm.io/gorm"
)

// WalletBalance represents a 12-word BIP39 wallet mnemonic generated from a LuckySix combination.
type WalletBalance struct {
	gorm.Model
	LuckySixID       uint       `gorm:"index:idx_wallet_luckysix;uniqueIndex:idx_wallet_mnemonic"`
	Mnemonic         string     `gorm:"type:text;uniqueIndex:idx_wallet_mnemonic"`
	CosmosAddress    string     `gorm:"column:cosmos_address"`
	Address          string     `gorm:"column:address"`
	Balance          string     `gorm:"column:balance"`        // Stored as string to handle large numbers
	CosmosBalance    string     `gorm:"column:cosmos_balance"` // Stored as string to handle large numbers
	BalanceUpdatedAt *time.Time `gorm:"column:balance_updated_at"`
	IsNotified       bool       `gorm:"column:is_notified;default:false"`
}
