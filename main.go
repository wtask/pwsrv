package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/wtask/pwsrv/internal/encryption/hasher"
	"github.com/wtask/pwsrv/internal/encryption/token"

	"github.com/wtask/pwsrv/internal/storage"

	"github.com/wtask/pwsrv/internal/storage/mysql"

	"github.com/wtask/pwsrv/internal/core"
)

const (
	AppConfigPathnameEnv = "PWSRV_CONFIG"
)

var (
	AppConfigPathname = ""
)

// start - launches given server to listen and serve in background;
// writes into fail channel an error (or nil) as reason of startup fall
// or termination.
func start(s *http.Server, fail chan<- error) {
	fmt.Printf("Starting server (%s) ...\n", s.Addr)
	go func() {
		fail <- s.ListenAndServe()
	}()
}

// waitStop - waiting for a value in the channel and stopping the server;
// writes shutdown error or nil into the fail channel.
func waitStop(s *http.Server, stop <-chan bool, fail chan<- error) {
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.SetKeepAlivesEnabled(false)
		fail <- s.Shutdown(ctx)
	}()
}

func storageFactory(cfg *Configuration) (storage.Interface, error) {
	var (
		storage storage.Interface
		err     error
	)
	switch cfg.Storage {
	case "mysql":
		storage, err = mysql.NewStorage(
			cfg.MySQL.DSN,
			mysql.WithTablePrefix("pwsrv_"),
			mysql.WithPasswordHasher(
				hasher.NewMD5DigestHasher(cfg.Secret.UserPassword),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("Storage factory: %s", err.Error())
		}
	default:
		return nil, fmt.Errorf("Storage factory: unsupported storage %q", cfg.Storage)
	}
	return storage, nil
}

func init() {
	descr := fmt.Sprintf(
		"Absolute file path to JSON config in case, if config location is not defined with %q environment's var.",
		AppConfigPathnameEnv,
	)
	AppConfigPathname, _ = os.LookupEnv(AppConfigPathnameEnv) // will initialize from env
	flag.StringVar(&AppConfigPathname, "config", AppConfigPathname, descr)
	flag.Parse()

	if AppConfigPathname == "" {
		fmt.Println("Startup error: config location is not defined.")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	cfg, err := loadJSONConfig(AppConfigPathname)
	if err != nil {
		fmt.Printf("Unable to load config (%s): %s\n", AppConfigPathname, err.Error())
		os.Exit(1)
	}

	storage, err := storageFactory(cfg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if storage == nil {
		fmt.Println("Storage interface was not properly initialized <nil>")
		os.Exit(1)
	}
	defer storage.Close()

	service, err := core.NewHTTPService(
		storage.CoreRepository(),
		token.NewMD5DigestBearer(cfg.Secret.AuthBearer, 1*time.Hour),
	)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port),
		Handler: core.Router(service),
	}

	once := sync.Once{}
	startFail := make(chan error, 1)
	stopFail := make(chan error, 1)
	stop := make(chan bool, 1)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	start(server, startFail)
	waitStop(server, stop, stopFail)

	for {
		select {
		case <-sig:
			stop <- true
		case err := <-startFail:
			if err != nil && err != http.ErrServerClosed {
				// any startup errors here...
				fmt.Printf("Server (%s) failed to run %q\n", server.Addr, err.Error())
				os.Exit(1)
			}
		case err := <-stopFail:
			if err != nil {
				fmt.Printf("Server (%s) stopped with an error %q\n", server.Addr, err.Error())
			} else {
				fmt.Printf("Server (%s) successfully stopped\n", server.Addr)
			}
			return
		default:
			once.Do(func() { fmt.Println("Server is running!") })
		}
	}
}
