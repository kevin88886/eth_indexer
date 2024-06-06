package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Event struct {
	ID          int64           `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	BlockNumber uint64          `gorm:"<-:create;column:block_number;type:bigint;index:idx_block_number;comment:'区块号'"`
	TxHash      string          `gorm:"<-:create;column:tx_hash;type:varchar(66);index:idx_hash;not null;comment:'交易哈希'"`
	Operate     string          `gorm:"<-:create;column:operate;type:varchar(20);index:idx_operate;not null;default:'';comment:'协议操作类型'"`
	Tick        string          `gorm:"<-:create;column:tick;type:varchar(64);index:idx_tick;not null;default:'';comment:'tick 名称'"`
	ETHFrom     string          `gorm:"<-:create;column:eth_from;type:varchar(42);index:idx_ierc_from;not null;comment:'IERC协议交易发起者'"`
	ETHTo       string          `gorm:"<-:create;column:eth_to;type:varchar(42);index:idx_ierc_to;not null;comment:'IERC协议交易接收者'"`
	IERCFrom    string          `gorm:"<-:create;column:ierc_from;type:varchar(42);index:idx_ierc_from;not null;comment:'IERC协议交易发起者'"`
	IERCTo      string          `gorm:"<-:create;column:ierc_to;type:varchar(42);index:idx_ierc_to;not null;comment:'IERC协议交易接收者'"`
	Amount      decimal.Decimal `gorm:"<-:create;column:amount;type:decimal(50,18);not null;default:0.000000000000000000;comment:'涉及的金额数量'"`
	Sign        string          `gorm:"<-:create;column:sign;type:varchar(256);index:idx_sign;not null;default:'';comment:'卖家签名'"`
	EventKind   uint8           `gorm:"<-:create;column:event_kind;type:tinyint;not null;default:0;comment:'事件类型'"`
	Event       []byte          `gorm:"<-:create;column:event_data;type:json;comment:'事件数据'"`
	ErrCode     int32           `gorm:"<-:create;column:err_code;type:int;index:idx_err_code;not null;default:0;comment:'处理结果状态码. 0: 表示成功'"`
	ErrReason   string          `gorm:"<-:create;column:err_reason;type:varchar(128);comment:'备注信息'"`
	EventAt     time.Time       `gorm:"<-:create;column:event_at;autoCreateTime:milli"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (i *Event) TableName() string {
	return "ierc_events"
}
