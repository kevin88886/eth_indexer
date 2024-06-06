package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type IERC20Balance struct {
	ID               int64           `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	Address          string          `gorm:"<-:create;column:address;type:varchar(42);uniqueIndex:uni_address_tick,priority:1;not null;default:'';"`
	Tick             string          `gorm:"<-:create;column:tick;type:varchar(64);uniqueIndex:uni_address_tick,priority:2;index:idx_tick;not null;default:'';comment:'tick'"`
	Available        decimal.Decimal `gorm:"column:available;type:decimal(50,18);not null;default:0.000000000000000000;comment:'可用额度'"`
	Freeze           decimal.Decimal `gorm:"column:freeze;type:decimal(50,18);not null;default:0.000000000000000000;comment:'冻结额度'"`
	Minted           decimal.Decimal `gorm:"column:minted;type:decimal(50,18);not null;default:0.000000000000000000;comment:'mint的数量'"`
	LastUpdatedBlock uint64          `gorm:"column:last_updated_block;type:bigint;comment:'区块号'"`
	CreatedAt        time.Time       `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        time.Time       `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (b *IERC20Balance) TableName() string {
	return "ierc_balances"
}
