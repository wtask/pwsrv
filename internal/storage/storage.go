package storage

import (
	"github.com/wtask/pwsrv/internal/core"
)

// Interface - common data storage access interface in accordance with the requirements of internal packages.
type Interface interface {
	CoreRepository() core.Repository
	Close() error
}
