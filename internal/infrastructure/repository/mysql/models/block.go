package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// 区块数据
type Block struct {
	ID               int64     `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	Number           uint64    `gorm:"<-:create;column:block_number;type:bigint;uniqueIndex:uni_block_number;comment:'区块号'"`
	Hash             string    `gorm:"<-:create;column:block_hash;type:varchar(66);not null;comment:'区块哈希'"`
	ParentHash       string    `gorm:"<-:create;column:parent_hash;type:varchar(66);not null;default:'';comment:'父哈希'"`
	TransactionCount int       `gorm:"<-:create;column:tx_count;type:bigint;index:idx_count;not null;default:0;comment:'当前区块包含的 Transaction 的数量'"`
	IsProcessed      bool      `gorm:"column:is_processed;type:int;not null;default:0;comment:'是否已处理. 0: 未处理; 1: 已处理'"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (b *Block) TableName() string {
	return "blocks"
}

// 交易数据
type Transaction struct {
	// 以太坊交易原始数据
	ID            int64           `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	BlockNumber   uint64          `gorm:"<-:create;column:block_number;type:bigint;uniqueIndex:uni_num_pos,priority:1;comment:'当前交易所属区块高度'"`
	PositionInTxs int64           `gorm:"<-:create;column:position;type:bigint;uniqueIndex:uni_num_pos,priority:2;comment:'当前交易在区块交易列表的位置'"`
	Hash          string          `gorm:"<-:create;column:hash;type:varchar(66);index:idx_hash;not null;comment:'区块哈希'"`
	From          string          `gorm:"<-:create;column:from;type:varchar(42);index:idx_from;not null;comment:'交易发起者'"`
	To            string          `gorm:"<-:create;column:to;type:varchar(42);index:idx_to;not null;comment:'交易接收者'"`
	Value         decimal.Decimal `gorm:"<-:create;column:value;type:decimal(65,0);not null;default:0;comment:'对应以太坊交易中的 Value, 即ETH的数量'"`
	Gas           decimal.Decimal `gorm:"<-:create;column:gas;type:decimal(65,0);not null;default:0;comment:'这笔交易消耗的gas'"`
	GasPrice      decimal.Decimal `gorm:"<-:create;column:gas_price;type:decimal(65,0);not null;default:0;"`
	Data          string          `gorm:"<-:create;column:data;type:MEDIUMBLOB;not null;comment:'当前交易的 input data'"`
	Nonce         uint64          `gorm:"<-:create;column:nonce;type:int;not null;comment:'Nonce'"`

	IsProcessed bool      `gorm:"column:is_processed;type:int;not null;default:0;comment:'是否已处理. 0: 未处理; 1: 已处理'"`
	Code        int32     `gorm:"column:code;type:int;not null;default:0;comment:'处理结果状态码'"`
	Remark      string    `gorm:"column:remark;type:varchar(128);comment:'备注信息'"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (t *Transaction) TableName() string {
	return "ierc_transactions"
}
