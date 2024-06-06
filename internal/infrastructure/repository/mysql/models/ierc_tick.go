package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type IERCTick struct {
	ID               int64           `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	Protocol         string          `gorm:"<-:create;column:protocol;type:varchar(20);not null;default:'';comment:'tick 所属协议'"`
	Tick             string          `gorm:"<-:create;column:tick;type:varchar(64);uniqueIndex:idx_tick;not null;default:'';comment:'tick 名称'"`
	Decimals         int64           `gorm:"<-:create;column:decimals;type:int;not null;default:0;comment:'tick 精度'"`
	Creator          string          `gorm:"<-:create;column:creator;type:varchar(64);not null;default:'';comment:'tick 创建者'"`
	MaxSupply        decimal.Decimal `gorm:"<-:create;column:max_supply;type:decimal(50,18);not null;default:0.000000000000000000;comment:'最大发行数量'"`
	Supply           decimal.Decimal `gorm:"column:supply;type:decimal(50,18);not null;default:0.000000000000000000;comment:'已发行数量'"`
	Detail           []byte          `gorm:"column:detail;type:json;comment:'详情数据'"`
	LastUpdatedBlock uint64          `gorm:"column:last_updated_block;type:bigint;comment:'区块号'"`
	CreatedAt        time.Time       `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        time.Time       `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (t *IERCTick) TableName() string {
	return "ierc_ticks"
}
