package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type Configuration struct {
	Server  ServerParams `json:"server"`
	Storage string       `json:"storage"`
	MySQL   MySQLParams  `json:"mysql"`
	Secret  SecretParams `json:"secret"`
}

type ServerParams struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type MySQLParams struct {
	DSN string `json:"dsn"`
}

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
	return &cfg, nil
}

func verifyConfig(cfg *Configuration) error {
	if cfg.Server.Port < 0 {
		return errors.New("config: server.Port value must be greater than or equal to zero")
	}
	if cfg.Storage != "mysql" {
		return errors.New("config: storage type supports only mysql")
	}
	if cfg.Storage == "mysql" {
		if cfg.MySQL.DSN == "" {
			return errors.New("config: mysql.dsn must not be empty")
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
