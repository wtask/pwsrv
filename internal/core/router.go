package core

import (
	"net/http"
	"strconv"

	"github.com/wtask/pwsrv/internal/core/reply"

	"github.com/wtask/pwsrv/internal/api"
	"github.com/wtask/pwsrv/internal/core/middleware"

	"github.com/gorilla/mux"
)

// Router - initializes router and returns http.Handler interface based on it.
func Router(service api.HTTPService) http.Handler {
	r := mux.NewRouter()
	r.Use(middleware.AuthorizationTryout(service.GetAuthBearer()))

	r.NewRoute().
		Path("/").
		Methods("OPTIONS").
		HandlerFunc(service.Options())

	r.NewRoute().
		Path("/login/").
		Methods("POST").
		HandlerFunc(service.Login())

	r.NewRoute().
		Path("/register/").
		Methods("POST").
		HandlerFunc(service.Register())

	{
		users := r.PathPrefix("/users/").Subrouter()
		users.Use(middleware.AuthorizationRequired())

		users.NewRoute().
			Path("/me/").
			Methods("GET").
			HandlerFunc(service.GetUserByAuth())

		users.NewRoute().
			Path("/{id:[0-9]+}/").
			Methods("GET").
			HandlerFunc(withID(service.GetUserByID))

		users.NewRoute().
			Path("/have-prefix/{prefix}/").
			Methods("GET").
			HandlerFunc(withString("prefix", service.UserListHavePrefix))
	}

	{
		transfers := r.PathPrefix("/money/transfers/").Subrouter()
		transfers.Use(middleware.AuthorizationRequired())

		transfers.NewRoute().
			Path("/").
			Methods("POST"). // create internal money transfer (IMT)
			HandlerFunc(service.CreateIMT())

		transfers.NewRoute().
			Path("/").
			Methods("GET"). // get list of IMT
			HandlerFunc(service.IMTCensoredList())

		transfers.NewRoute().
			Path("/{id:[0-9]+}/").
			Methods("POST"). // create new IMT based on given ID
			HandlerFunc(withID(service.RepeatIMTByID))

		transfers.NewRoute().
			Path("/{id:[0-9]+}/").
			Methods("GET"). // get IMT details for given ID
			HandlerFunc(withID(service.GetIMTCensoredByID))
	}

	return r
}

// withID - is a helper for using service handler which required
// unsigned integer argument which is the named (id) part of several routes.
func withID(adaptee func(uint64) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
		if err != nil || id < 1 {
			reply.BadRequest("Bad request")(w, r)
			return
		}
		adaptee(id)(w, r)
	}
}

func withString(varName string, adaptee func(string) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		value, ok := mux.Vars(r)[varName]
		if !ok || value == "" {
			reply.BadRequest("Bad request")(w, r)
			return
		}
		adaptee(value)(w, r)
	}
}
