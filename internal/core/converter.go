package core

import (
	"github.com/wtask/pwsrv/internal/api"
	"github.com/wtask/pwsrv/internal/model"
)

// censorInternalTransfer - converts model into response item format
func censorInternalTransfer(memberID uint64, t *model.InternalTransfer) *api.IMTCensored {
	if t == nil {
		return nil
	}
	// research credit/debit from member point of view
	isCredit := memberID == t.RecipientID
	userID := t.UserID
	balanceBefore := t.RecipientBalanceBefore
	balanceAfter := t.RecipientBalanceAfter
	sum := t.Sum
	if !isCredit {
		userID = t.RecipientID
		balanceBefore = t.UserBalanceBefore
		balanceAfter = t.UserBalanceAfter
		sum = -t.Sum
	}
	c := &api.IMTCensored{
		ID:            t.ID,
		Date:          t.CreatedAt,
		IsCredit:      isCredit,
		Sum:           sum,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		UserID:        userID,
	}

	return c
}
