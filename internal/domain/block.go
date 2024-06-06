package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kevin88886/eth_indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
)

type BlockHeader struct {
	Number     uint64
	Hash       string
	ParentHash string
}

func (b *BlockHeader) String() string {
	if b == nil {
		return "0"
	} else {
		return strconv.FormatUint(b.Number, 10)
	}
}

type BlockHandleStatus struct {
	LatestBlock      *BlockHeader // 远程节点最新区块, 来自节点
	LastIndexedBlock *BlockHeader // 当前已经索引到的区块, 来自数据库
	LastSyncBlock    *BlockHeader // 当前已经同步完成的区块, 来自数据库
}

func (b *BlockHandleStatus) String() string {
	if b == nil {
		return "nil"
	}

	return fmt.Sprintf("latestBlock: %s, indexedBlock: %s, syncBlock: %s", b.LatestBlock, b.LastIndexedBlock, b.LastSyncBlock)
}

type Block struct {
	Number           uint64         // 区块号
	ParentHash       string         // 父区块哈希
	Hash             string         // 区块哈希
	TransactionCount int            // 当前区块包含的 Transaction 的数量
	Transactions     []*Transaction // 当前区块中的交易数据
	IsProcessed      bool           // 是否已处理
	CreatedAt        time.Time      // 区块索引时间
	UpdatedAt        time.Time      // 区块处理时间
}

func (b *Block) Header() *BlockHeader {
	return &BlockHeader{
		Number:     b.Number,
		Hash:       b.Hash,
		ParentHash: b.ParentHash,
	}
}

type Transaction struct {
	// 以太坊交易原始数据
	BlockNumber   uint64          // 当前交易所属区块号
	PositionInTxs int64           // 当前交易在区块交易列表的位置
	Hash          string          // 交易hash
	From          string          // 交易发起者
	To            string          // 交易接收者
	TxData        string          // 当前交易的 input data
	TxValue       decimal.Decimal // 对应以太坊交易中的 Value, types.Transaction.Value
	Gas           decimal.Decimal
	GasPrice      decimal.Decimal
	Nonce         uint64

	// 交易处理状态
	IsProcessed bool   // 是否已处理
	Code        int32  // 处理结果
	Remark      string // 备注
	CreatedAt   time.Time
	UpdatedAt   time.Time

	IERCTransaction protocol.IERCTransaction
}
