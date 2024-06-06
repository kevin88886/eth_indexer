package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type StakingPool struct {
	ID               int64     `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	Pool             string    `gorm:"<-:create;column:pool;type:varchar(42);uniqueIndex:uni_pool,priority:1;comment:'质押池地址'"`
	PoolID           uint64    `gorm:"<-:create;column:pool_id;type:bigint;uniqueIndex:uni_pool,priority:2;comment:'质押池ID'"`
	Name             string    `gorm:"column:name;type:varchar(64);comment:'质押池名称'"`
	Owner            string    `gorm:"column:owner;type:varchar(42);index:id_owner;not null;comment:'池子管理员'"`
	Data             []byte    `gorm:"column:data;type:json;comment:'质押池的相关数据'"`
	LastUpdatedBlock uint64    `gorm:"column:last_updated_block;type:bigint;comment:'更新时的区块高度'"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (t *StakingPool) TableName() string {
	return "staking_pools"
}

type StakingPosition struct {
	ID               int64           `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	Pool             string          `gorm:"<-:create;column:pool;type:varchar(42);uniqueIndex:uni_pool_staker,priority:1;not null;comment:'质押池地址'"`
	PoolID           uint64          `gorm:"<-:create;column:pool_id;type:bigint;uniqueIndex:uni_pool_staker,priority:2;comment:'池子ID'"`
	Staker           string          `gorm:"<-:create;column:staker;type:varchar(42);uniqueIndex:uni_pool_staker,priority:3;not null;comment:'质押者地址'"`
	AccRewards       decimal.Decimal `gorm:"column:acc_rewards;type:decimal(50,18);not null;default:0.000000000000000000;comment:'用户到上一个奖励区块时，累积了多少奖励'"`
	Debt             decimal.Decimal `gorm:"column:debt;type:decimal(50,18);not null;default:0.000000000000000000;comment:'用于一共使用了多少奖励'"`
	RewardsPerBlock  decimal.Decimal `gorm:"column:rewards_per_block;type:decimal(50,18);not null;default:0.000000000000000000;comment:'用户每个块的奖励'"`
	LastRewardBlock  uint64          `gorm:"column:last_reward_block;type:bigint;comment:'用户上一次获得奖励的区块'"`
	LastUpdatedBlock uint64          `gorm:"column:last_updated_block;type:bigint;comment:'用户上一次更新仓位信息的区块'"`
	Amounts          []byte          `gorm:"column:staker_amounts;type:json;comment:'质押的Tick数量'"`
	CreatedAt        time.Time       `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        time.Time       `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (t *StakingPosition) TableName() string {
	return "staking_positions"
}

type StakingBalance struct {
	ID          int64           `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	Staker      string          `gorm:"<-:create;column:staker;type:varchar(42);uniqueIndex:uni_staker_pool_tick,priority:1;not null;comment:'质押者地址'"`
	Pool        string          `gorm:"<-:create;column:pool;type:varchar(42);uniqueIndex:uni_staker_pool_tick,priority:2;index:idx_pool;not null;comment:'质押池地址'"`
	PoolID      uint64          `gorm:"<-:create;column:pool_id;type:bigint;uniqueIndex:uni_staker_pool_tick,priority:3;index:idx_pool_id;comment:'池子ID'"`
	Tick        string          `gorm:"<-:create;column:tick;type:varchar(64);uniqueIndex:uni_staker_pool_tick,priority:4;not null;default:'';comment:'质押的Tick'"`
	Amount      decimal.Decimal `gorm:"column:amount;type:decimal(50,18);not null;default:0.000000000000000000;comment:'质押的数量'"`
	BlockNumber uint64          `gorm:"column:block_number;type:bigint;comment:'更新时的区块高度'"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (t *StakingBalance) TableName() string {
	return "staking_balances"
}
