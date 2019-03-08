package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

// Configuration - application runtime parameters
type Configuration struct {
	Server      ServerParams `json:"server"`
	DSN         string       `json:"dsn"`
	StorageType string       `json:-`
	MySQL       MySQLOptions `json:"mysql"`
	Secret      SecretParams `json:"secret"`
}

// ServerParams - application server parameters
type ServerParams struct {
	Address string `json:"address"`
	Port    int    `json:"port,string"`
}

// MySQLOptions - mysql connection options
type MySQLOptions struct {
	ParseTime bool   `json:"parse_time"`
	Timeout   string `json:"connect_timeout"`
}

// SecretParams - strings used to generate sensitive data
type SecretParams struct {
	UserPassword string `json:"user_password"`
	AuthBearer   string `json:"auth_bearer"`
}

func loadJSONConfig(filepath string) (*Configuration, error) {
	src := []byte{}
	src, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	cfg := Configuration{}
	if err = json.Unmarshal(src, &cfg); err != nil {
		return nil, err
	}

	if i := strings.Index(cfg.DSN, "://"); i > 0 {
		// init storage type
		cfg.StorageType = strings.ToLower(cfg.DSN[:i])
	}

	if err = verifyConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func verifyConfig(cfg *Configuration) error {
	if cfg.Server.Port < 0 {
		return errors.New("config: server.Port value must be greater than or equal to zero")
	}

	if cfg.StorageType != "mysql" {
		return errors.New("config: invalid dsn, only mysql:// connection is supported")
	}
	if cfg.StorageType == "mysql" {
		// exclude any mysql connection options given as dsn's query string
		if strings.Index(cfg.DSN, "?") != -1 {
			return errors.New("config: mysql dsn must not contain query string for options")
		}
	}

	if cfg.Secret.UserPassword == "" {
		return errors.New("config: secret.user_password must not be empty")
	}
	if cfg.Secret.AuthBearer == "" {
		return errors.New("config: secret.auth_bearer must not be empty")
	}

	return nil
}

func (m MySQLOptions) String() string {
	s := fmt.Sprintf("parseTime=%t", m.ParseTime)
	if m.Timeout != "" {
		s += fmt.Sprintf("&timeout=%s", m.Timeout)
	}
	return s
}
