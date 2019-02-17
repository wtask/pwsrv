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

func WithDSN(dsn string) storageOption {
	if dsn == "" {
		panic(errors.New("mysql.WithDSN: DSN is empty"))
	}
	return func(s *mysqlstorage) {
		s.dsn = strings.TrimPrefix(dsn, `mysql://`)
	}
}

func WithTablePrefix(prefix string) storageOption {
	return func(s *mysqlstorage) {
		s.tablePrefix = prefix
	}
}

func WithPasswordHasher(h hasher.StringHasher) storageOption {
	if h == nil {
		panic(errors.New("mysql.WithPasswordHasher: string hasher is nil"))
	}
	return func(s *mysqlstorage) {
		s.passwordHasher = h
	}
}

func (s *mysqlstorage) alter(options ...storageOption) *mysqlstorage {
	if s == nil {
		return nil
	}
	for _, o := range options {
		if o != nil {
			o(s)
		}
	}
	return s
}

func NewStorage(options ...storageOption) (storage.Interface, error) {
	s := (&mysqlstorage{}).alter(options...)
	if s.dsn == "" {
		return nil, errors.New("mysql.NewStorage: DSN is empty")
	}
	if s.passwordHasher == nil {
		return nil, errors.New("mysql.NewStorage: password hasher is nil")
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

	return s, nil
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
