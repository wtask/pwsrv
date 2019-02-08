package mysql

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/wtask/pwsrv/internal/model"
)

func (s *mysqlstorage) GetUserByID(userID uint64) (*model.User, error) {
	user := model.User{}
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("mysql.GetUserByID: %s", err.Error())
	}
	return &user, nil
}

func (s *mysqlstorage) GetUserByEmail(address string) (*model.User, error) {
	u := &model.User{Email: address}
	if err := s.db.Where(u).First(u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("mysql.GetUserByEmail: %s", err.Error())
	}
	return u, nil
}

func (s *mysqlstorage) GetUserByEmailAndPassword(address, password string) (*model.User, error) {
	user, err := s.GetUserByEmail(address)
	if err != nil {
		return nil, fmt.Errorf("mysql.GetUserByEmailAndPassword: %s", err.Error())
	}
	if user == nil ||
		user.PHash != s.passwordHasher.Hash(password) {
		return nil, nil
	}
	return user, nil
}

func (s *mysqlstorage) CreateUser(u model.User, password string) (*model.User, error) {
	if u.ID != 0 ||
		u.Email == "" ||
		u.Name == "" {
		return nil, errors.New("mysql.CreateUser: existed ID or required field is empty")
	}
	if password == "" {
		return nil, errors.New("mysql.CreateUser: required password is empty")
	}
	user, err := s.GetUserByEmail(u.Email)
	if err != nil {
		return nil, fmt.Errorf("mysql.CreateUser: %s", err.Error())
	}
	if user != nil {
		return nil, errors.New("mysql.CreateUser: already exists")
	}
	u.PHash = s.passwordHasher.Hash(password)
	if err := s.db.Create(&u).Error; err != nil {
		return nil, fmt.Errorf("mysql.CreateUser: %s", err.Error())
	}
	return &u, nil
}

func (s *mysqlstorage) FindUsersHavePrefix(prefix string, limit int) ([]model.User, error) {
	if prefix == "" || limit <= 0 {
		return nil, nil
	}
	users := []model.User{}
	prefix = strings.NewReplacer(`%`, `\%`, `_`, `\_`).Replace(prefix)
	err := s.db.Where("name LIKE ? OR email LIKE ?", prefix+"%", prefix+"%").Limit(limit).Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("mysql.FindUsersHavePrefix: %s", err.Error())
	}
	return users, nil
}

func (s *mysqlstorage) CreateInternalTransfer(userID, recipientID uint64, sum float64) (*model.InternalTransfer, error) {
	u, r := model.User{}, model.User{}
	tx := s.db.Begin()
	err := tx.Where("id = ? and balance - ? > 0", userID, sum).First(&u).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("mysql.CreateInternalTransfer: %s", err.Error())
	}
	if u.ID == 0 {
		tx.Rollback()
		return nil, fmt.Errorf("mysql.CreateInternalTransfer: user #%d not found or insufficient funds", userID)
	}
	if err = tx.First(&r, recipientID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("mysql.CreateInternalTransfer: recipient #%d not found", recipientID)
		}
		return nil, fmt.Errorf("mysql.CreateInternalTransfer: %s", err.Error())
	}

	uBalance, rBalance := u.Balance, r.Balance
	// transfer
	// debit first
	err = tx.Model(&u).Where("balance - ? > 0", sum).UpdateColumn("balance", gorm.Expr("balance - ?", sum)).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("mysql.CreateInternalTransfer: %s", err.Error())
	}
	if err = tx.First(&u, userID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("mysql.CreateInternalTransfer: user #%d is missing", recipientID)
		}
		return nil, fmt.Errorf("mysql.CreateInternalTransfer: %s", err.Error())
	}
	if u.Balance < 0 {
		tx.Rollback()
		return nil, fmt.Errorf("mysql.CreateInternalTransfer: debit processing causes insufficient funds")
	}
	// credit
	err = tx.Model(&r).UpdateColumn("balance", gorm.Expr("balance + ?", sum)).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("mysql.CreateInternalTransfer: %s", err.Error())
	}
	if err = tx.First(&r, r.ID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("mysql.CreateInternalTransfer: recipient #%d is missing", recipientID)
		}
		return nil, fmt.Errorf("mysql.CreateInternalTransfer: %s", err.Error())
	}
	// transaction log
	itm := model.InternalTransfer{
		CreatedAt:              time.Now().UTC(),
		UserID:                 u.ID,
		RecipientID:            r.ID,
		Sum:                    sum,
		UserBalanceBefore:      uBalance,
		UserBalanceAfter:       u.Balance,
		RecipientBalanceBefore: rBalance,
		RecipientBalanceAfter:  r.Balance,
	}
	if err = tx.Create(&itm).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("mysql.CreateInternalTransfer: %s", err.Error())
	}
	if itm.ID == 0 {
		tx.Rollback()
		return nil, fmt.Errorf("mysql.CreateInternalTransfer: cannot finish transaction (#%d, %f) -> #%d", u.ID, sum, r.ID)
	}
	tx.Commit()

	return &itm, nil
}

func (s *mysqlstorage) GetInternalTransferByID(transferID uint64) (*model.InternalTransfer, error) {
	t := model.InternalTransfer{}
	if err := s.db.First(&t, transferID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("mysql.GetInternalTransferByID: %s", err.Error())
	}
	return &t, nil
}

func (s *mysqlstorage) RepeatInternalTransfer(transferID uint64) (*model.InternalTransfer, error) {
	t, err := s.GetInternalTransferByID(transferID)
	if err != nil {
		return nil, fmt.Errorf("mysql.RepeatInternalTransfer: %s", err.Error())
	}
	if t == nil {
		return nil, fmt.Errorf("mysql.RepeatInternalTransfer: transfer not found #%d", transferID)
	}
	return s.CreateInternalTransfer(t.UserID, t.RecipientID, t.Sum)
}

func (s *mysqlstorage) FindLastInternalTransfers(userID uint64, limit int) ([]model.InternalTransfer, error) {
	if limit <= 0 {
		return nil, nil
	}
	transfers := []model.InternalTransfer{}
	err := s.db.
		Where("user_id = ? OR recipient_id = ?", userID, userID).
		Order("id DESC").
		Limit(limit).
		Find(&transfers).
		Error
	if err != nil {
		return nil, fmt.Errorf("mysql.FindLastInternalTransfers: %s", err.Error())
	}
	return transfers, nil
}
