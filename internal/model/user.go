package model

import (
	"time"
)

// User - user entity
type User struct {
	ID        uint64    `gorm:"primary_key" json:"id,string"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:current_timestamp on update current_timestamp" json:"-"`
	Role      UserRole  `gorm:"not null;default:'1'" json:"role,string"`
	Email     string    `gorm:"not null;unique_index" json:"email"`
	Name      string    `gorm:"not null;index" json:"name"`
	PHash     string    `gorm:"not null" json:"-"`
	Balance   float64   `gorm:"not null;default:'0'" json:"balance,string"`
}

// UserRole - simple roles enumeration
type UserRole byte

const (
	_ UserRole = iota
	// RoleRegular - new, not trusted user
	RoleRegular
	// RoleTrusted - verified user
	RoleTrusted
)
