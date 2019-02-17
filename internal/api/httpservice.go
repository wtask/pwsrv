package api

import (
	"net/http"
	"time"

	"github.com/wtask/pwsrv/internal/model"
)

type (
	HTTPService interface {
		Options() http.HandlerFunc // not impl
		Login() http.HandlerFunc
		Register() http.HandlerFunc
		UserListHavePrefix(prefix string) http.HandlerFunc
		GetUserByID(id uint64) http.HandlerFunc
		GetUserByAuth() http.HandlerFunc
		IMTCensoredList() http.HandlerFunc
		GetIMTCensoredByID(id uint64) http.HandlerFunc
		CreateIMT() http.HandlerFunc
		RepeatIMTByID(id uint64) http.HandlerFunc
	}
)

type (
	// ErrorResponse - common error response
	ErrorResponse struct {
		// Error - flag, always true for errors and may present as false in successfull response
		Error bool `json:"error"`
		// Message - error message
		Message string `json:"message,omitempty"`
	}

	// LoginResponse - successfull Login response
	LoginResponse struct {
		Auth string `json:"auth"`
	}

	// RegisterResponse - successfull Register response
	RegisterResponse struct {
		Auth string `json:"auth"`
	}

	// GetUserResponse - successfull GetUserByXXX response
	GetUserResponse struct {
		User *model.User `json:"user"`
	}

	// UserListResponse - successfull UserListXXX response
	UserListResponse struct {
		Users []model.User `json:"users"`
	}

	UserListHavePrefix = UserListResponse

	// IDResponse - successfull response of methods, which returns single ID
	IDResponse struct {
		ID uint64 `json:"id,string"`
	}

	// CreateIMTResponse - successfull CreateIMT response
	CreateIMTResponse = IDResponse
	// RepeatIMTResponse - successfull RepeatIMT response
	RepeatIMTResponse = IDResponse

	// IMTCensored - censored internal money transfer data
	IMTCensored struct {
		ID            uint64    `json:"id,string"`
		Date          time.Time `json:"date"`
		IsCredit      bool      `json:"is_credit"`
		Sum           float64   `json:"sum,string"`
		BalanceBefore float64   `json:"balance_before,string"`
		BalanceAfter  float64   `json:"balance_after,string"`
		UserID        uint64    `json:"user_id,string"`
	}

	// GetIMTCensoredResponse - successfull GetIMTCensoredByXXX response
	GetIMTCensoredResponse struct {
		Transaction *IMTCensored `json:"transaction"`
	}
	// IMTCensoredListResponse - successfull IMTCensoredList response
	IMTCensoredListResponse struct {
		Transactions []*IMTCensored `json:"transactions"`
	}
)
