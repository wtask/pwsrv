package core

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/wtask/pwsrv/internal/core/middleware"

	"github.com/wtask/pwsrv/internal/core/reply"

	"github.com/wtask/pwsrv/pkg/email"

	"github.com/wtask/pwsrv/internal/model"

	"github.com/wtask/pwsrv/internal/api"
)

type (
	Repository interface {
		GetUserByID(userID uint64) (*model.User, error)
		GetUserByEmailAndPassword(address, password string) (*model.User, error)
		CreateUser(user model.User, password string) (*model.User, error)
		FindUsersHavePrefix(prefix string, limit int) ([]model.User, error)
		CreateInternalTransfer(userID, recipientID uint64, sum float64) (*model.InternalTransfer, error)
		RepeatInternalTransfer(transferID uint64) (*model.InternalTransfer, error)
		GetInternalTransferByID(transferID uint64) (*model.InternalTransfer, error)
		FindLastInternalTransfers(userID uint64, limit int) ([]model.InternalTransfer, error)
	}

	TokenProvider interface {
		NewToken(userID uint64) string
	}
)

// service - HTTPService interface implementation
type service struct {
	r Repository
	b TokenProvider
}

// NewHTTPService - builds api.HTTPService interface implementation.
func NewHTTPService(r Repository, b TokenProvider) (api.HTTPService, error) {
	if r == nil {
		return nil, errors.New("NewHTTPService(): Repository is nil")
	}
	if b == nil {
		return nil, errors.New("NewHTTPService(): TokenProvider is nil")
	}
	return &service{
		r: r,
		b: b,
	}, nil
}

const (
	MinPasswordLen = 5
)

func (s *service) Options() http.HandlerFunc {
	return reply.ServiceUnavailable()
}

func (s *service) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			reply.BadRequest("Can't parse post data")(w, r)
			return
		}
		login, password := r.Form.Get("login"), r.Form.Get("password")
		if login == "" || password == "" {
			reply.BadRequest("Required both non-empty login and password")(w, r)
			return
		}
		a := email.NewAddress(login)
		if !a.IsValid() || a.UserName() != "" {
			reply.BadRequest("Invalid login, valid email address expected")(w, r)
			return
		}
		user, err := s.r.GetUserByEmailAndPassword(a.Get(), password)
		if err != nil {
			// TODO log repo error
			reply.InternalServerError("Cannot complete request now")(w, r)
			return
		}
		if user == nil {
			reply.Conflict("Unable to authorize with given credentials")(w, r)
			return

		}
		token := s.b.NewToken(user.ID)
		if token == "" {
			// TODO log bearer error
			reply.InternalServerError("Cannot prepare authorization token now")(w, r)
			return
		}

		reply.OK(&api.LoginResponse{fmt.Sprintf("Bearer %s", token)})(w, r)
	}
}

func (s *service) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			reply.BadRequest("Can't parse post data")(w, r)
			return
		}
		login, password, name := r.Form.Get("email"), r.Form.Get("password"), r.Form.Get("name")
		if login == "" {
			reply.BadRequest("Required login is empty")(w, r)
			return
		}
		a := email.NewAddress(login)
		if !a.IsValid() || a.UserName() != "" {
			reply.BadRequest("Invalid login, email address required")(w, r)
			return
		}
		if len(password) < MinPasswordLen {
			reply.BadRequest(
				fmt.Sprintf("Password length must be %d or greater", MinPasswordLen),
			)(w, r)
			return
		}
		if name == "" {
			reply.BadRequest("Required name is empty")(w, r)
			return
		}
		user, err := s.r.CreateUser(
			model.User{
				Email:   login,
				Name:    name,
				Role:    model.RoleRegular,
				Balance: 500.0,
			},
			password,
		)
		if err != nil || user == nil {
			reply.InternalServerError("Cannot complete request")(w, r)
			return
		}
		token := s.b.NewToken(user.ID)
		reply.OK(&api.RegisterResponse{fmt.Sprintf("Bearer %s", token)})(w, r)
	}
}

func (s *service) UserListHavePrefix(prefix string) http.HandlerFunc {
	// return reply.ServiceUnavailable()
	return func(w http.ResponseWriter, r *http.Request) {
		authUser, ok := s.authorize(r)
		if !ok {
			reply.Unauthorized()(w, r)
			return
		}
		if authUser.Role < model.RoleTrusted {
			reply.Forbidden("Insufficient authority to complete request")(w, r)
			return
		}
		if len(prefix) < 3 {
			// depends on client preference
			// reply error or empty list
			reply.Conflict("Prefix too short")(w, r)
			return
		}

		list, err := s.r.FindUsersHavePrefix(prefix, 100)
		if err != nil {
			reply.InternalServerError("Cannot complete request now")(w, r)
			return
		}
		reply.OK(&api.UserListHavePrefix{Users: list})(w, r)
	}
}

func (s *service) GetUserByID(id uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authUser, ok := s.authorize(r)
		if !ok {
			reply.Unauthorized()(w, r)
			return
		}
		if id == 0 {
			reply.BadRequest("Invalid user ID")(w, r)
			return
		}
		if authUser.ID != id {
			if authUser.Role < model.RoleTrusted {
				reply.Forbidden("Insufficient authority to complete request")(w, r)
				return
			}
			user, err := s.r.GetUserByID(id)
			if err != nil {
				reply.InternalServerError("Cannot complete request now")(w, r)
				return
			}
			if user == nil {
				reply.Conflict("User not found")(w, r)
				return
			}
			reply.OK(&api.GetUserResponse{User: user})(w, r)
			return
		}
		reply.OK(&api.GetUserResponse{User: authUser})(w, r)
	}
}

func (s *service) GetUserByAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authUser, ok := s.authorize(r)
		if !ok {
			reply.Unauthorized()(w, r)
			return
		}
		reply.OK(&api.GetUserResponse{User: authUser})(w, r)
	}
}

func (s *service) IMTCensoredList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authUser, ok := s.authorize(r)
		if !ok {
			reply.Unauthorized()(w, r)
			return
		}
		transfers, err := s.r.FindLastInternalTransfers(authUser.ID, 100)
		if err != nil {
			reply.InternalServerError("Cannot complete request")(w, r)
			return
		}
		censored := make([]*api.IMTCensored, len(transfers))
		for i := range transfers {
			censored[i] = censorInternalTransfer(authUser.ID, &transfers[i])
		}
		reply.OK(&api.IMTCensoredListResponse{Transactions: censored})(w, r)
	}
}

func (s *service) GetIMTCensoredByID(id uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authUser, ok := s.authorize(r)
		if !ok {
			reply.Unauthorized()(w, r)
			return
		}
		transfer, err := s.r.GetInternalTransferByID(id)
		if err != nil {
			reply.InternalServerError("Cannot complete request")(w, r)
			return
		}
		if transfer == nil {
			reply.InternalServerError("Transfer not found")(w, r)
			return
		}
		if transfer.UserID != authUser.ID && transfer.RecipientID != authUser.ID {
			reply.Forbidden("Insufficient authority to complete request")(w, r)
			return
		}
		reply.OK(&api.GetIMTCensoredResponse{
			Transaction: censorInternalTransfer(authUser.ID, transfer),
		})(w, r)
	}
}

func (s *service) CreateIMT() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authUser, ok := s.authorize(r)
		if !ok {
			reply.Unauthorized()(w, r)
			return
		}
		if authUser.Role < model.RoleTrusted {
			reply.Forbidden("Insufficient authority to complete request")(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			reply.BadRequest("Can't parse post data")(w, r)
			return
		}
		recipientID, err := strconv.ParseUint(r.Form.Get("recipient_id"), 10, 64)
		if err != nil || recipientID == 0 || recipientID == authUser.ID {
			reply.BadRequest("Invalid recipient ID")(w, r)
			return
		}
		if recipient, _ := s.r.GetUserByID(recipientID); recipient == nil {
			reply.Conflict("Recipient not found")(w, r)
			return
		}

		sum, err := strconv.ParseFloat(r.Form.Get("sum"), 10)
		if err != nil || sum <= 0.0 {
			reply.BadRequest("Incorrect sum")(w, r)
			return
		}

		if authUser.Balance-sum < 0.0 {
			reply.Conflict(fmt.Sprintf("Insufficient funds (%f)", authUser.Balance))(w, r)
			return
		}

		transfer, err := s.r.CreateInternalTransfer(authUser.ID, recipientID, sum)
		if err != nil {
			reply.Conflict(fmt.Sprintf("Cannot transfer money to #%d", recipientID))(w, r)
			return
		}
		reply.OK(&api.CreateIMTResponse{ID: transfer.ID})(w, r)
	}
}

func (s *service) RepeatIMTByID(id uint64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authUser, ok := s.authorize(r)
		if !ok {
			reply.Unauthorized()(w, r)
			return
		}
		if id == 0 {
			reply.BadRequest("Invalid transfer ID")(w, r)
			return
		}
		transfer, err := s.r.GetInternalTransferByID(id)
		if err != nil {
			reply.InternalServerError("Cannot complete request now")(w, r)
			return
		}
		if transfer == nil {
			reply.Conflict("Transfer not found")(w, r)
			return
		}
		if transfer.UserID != authUser.ID {
			reply.Forbidden("Insufficient authority to repeat transfer")(w, r)
			return
		}
		newTransfer, err := s.r.RepeatInternalTransfer(transfer.ID)
		if err != nil {
			reply.Conflict(fmt.Sprintf("Cannot repeat transfer #%d", id))(w, r)
			return
		}
		reply.OK(&api.RepeatIMTResponse{ID: newTransfer.ID})(w, r)
	}
}

func (s *service) authorize(r *http.Request) (auth *model.User, ok bool) {
	userID, ok := middleware.DiscoverUserID(r)
	if !ok || userID == 0 {
		return nil, false
	}
	user, err := s.r.GetUserByID(userID)
	if err != nil {
		return nil, false
	}
	return user, true
}
