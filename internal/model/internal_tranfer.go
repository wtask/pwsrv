package model

import (
	"time"
)

// InternalTransfer - log of internal transfers
type InternalTransfer struct {
	ID          uint64    `gorm:"primary_key" json:"id,string"`
	CreatedAt   time.Time `gorm:"not null;default:current_timestamp" json:"created_at"`
	UserID      uint64    `gorm:"not null;index" json:"user_id,string"`
	RecipientID uint64    `gorm:"not null;index" json:"recipient_id,string"`
	Sum         float64   `gorm:"not null" json:"sum,string"`
	// balances have `omitempty` because may are 2 kind of replies:
	// sender must not see receiver's balance
	// receiver must not see sender's balance
	UserBalanceBefore      float64 `gorm:"not null" json:"user_balance_before,string,omitempty"`
	UserBalanceAfter       float64 `gorm:"not null" json:"user_balance_after,string,omitempty"`
	RecipientBalanceBefore float64 `gorm:"not null" json:"recipient_balance_before,string,omitempty"`
	RecipientBalanceAfter  float64 `gorm:"not null" json:"recipient_balance_after,string,omitempty"`
}
