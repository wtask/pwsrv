package mysql

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	// "go-sql-driver/mysql" initialization via gorm wrapper
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/wtask/pwsrv/internal/core"
	"github.com/wtask/pwsrv/internal/encryption/hasher"
	"github.com/wtask/pwsrv/internal/model"
	"github.com/wtask/pwsrv/internal/storage"
)

type mysqlstorage struct {
	db             *gorm.DB
	dsn            string
	tablePrefix    string
	passwordHasher hasher.StringHasher
}

type storageOption func(*mysqlstorage)

func WithTablePrefix(prefix string) storageOption {
	return func(s *mysqlstorage) {
		s.tablePrefix = prefix
	}
}

func WithPasswordHasher(h hasher.StringHasher) storageOption {
	return func(s *mysqlstorage) {
		s.passwordHasher = h
	}
}

func NewStorage(dsn string, options ...storageOption) (storage.Interface, error) {
	s := mysqlstorage{
		dsn: strings.TrimPrefix(dsn, `mysql://`),
		// insecure hasher without secret
		passwordHasher: hasher.NewMD5DigestHasher(""),
	}
	if s.dsn == "" {
		return nil, errors.New("mysql.NewStorage(): DSN required")
	}
	for _, o := range options {
		o(&s)
	}
	if s.tablePrefix != "" {
		gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
			return s.tablePrefix + defaultTableName
		}
	}

	db, err := gorm.Open("mysql", s.dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql.NewStorage(): %s", err.Error())
	}
	s.db = db

	s.db.SingularTable(true) // do not use plural form of table name
	err = s.db.
		Set("gorm:table_options", "COLLATE='utf8_general_ci' ENGINE=InnoDB").
		AutoMigrate(&model.User{}, &model.InternalTransfer{}).
		Error
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *mysqlstorage) CoreRepository() core.Repository {
	if s.db == nil {
		return nil
	}
	return s
}

func (s *mysqlstorage) Close() error {
	if s.db == nil {
		return errors.New("mysql.Close(): storage is not initialized")
	}
	return s.db.Close()
}
