package tick

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kevin88886/eth_indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
)

type IERC20Tick struct {
	ID                 int64             `json:"id,omitempty"`
	Protocol           protocol.Protocol `json:"protocol,omitempty"` // tick 所属协议
	Tick               string            `json:"tick,omitempty"`     // tick 名称
	MaxSupply          decimal.Decimal   `json:"max_supply"`         // 最大发行量
	Supply             decimal.Decimal   `json:"supply"`             // 已发行数量
	Decimals           int64             `json:"decimals,omitempty"` // tick 精度
	Limit              decimal.Decimal   `json:"limit"`              // 一笔交易最大mint数量
	WalletLimit        decimal.Decimal   `json:"wallet_limit"`       // 一个地址最多mint数量
	WorkC              string            `json:"work_c,omitempty"`   // mint 难度. 0x0000
	Creator            string            `json:"creator,omitempty"`  // tick 创建者
	LastUpdatedAtBlock uint64            `json:"updated_at_block"`   // 最后更新于哪个区块
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

func NewTickFromDeployCommand(command *protocol.DeployCommand) *IERC20Tick {
	maxSupply := command.MaxSupply
	// 对以下两个币做特殊处理
	switch command.Tick {
	case "ierc-m4":
		maxSupply = decimal.NewFromInt(21_016_000)
	case "ierc-m5":
		maxSupply = decimal.NewFromInt(24_874_936)
	}

	return &IERC20Tick{
		ID:                 0,
		Protocol:           command.Protocol,
		Tick:               command.Tick,
		MaxSupply:          maxSupply,
		Supply:             decimal.Zero,
		Decimals:           command.Decimals,
		Limit:              command.MintLimitOfSingleTx,
		WalletLimit:        command.MintLimitOfWallet,
		WorkC:              command.Workc,
		Creator:            command.From,
		LastUpdatedAtBlock: command.BlockNumber,
		CreatedAt:          command.EventAt,
		UpdatedAt:          command.EventAt,
	}
}

func (t *IERC20Tick) GetID() int64                   { return t.ID }
func (t *IERC20Tick) GetProtocol() protocol.Protocol { return t.Protocol }
func (t *IERC20Tick) GetName() string                { return t.Tick }
func (t *IERC20Tick) LastUpdatedBlock() uint64       { return t.LastUpdatedAtBlock }

// 验证hash
func (t *IERC20Tick) ValidateHash(hash string) error {

	if len(t.WorkC) != 0 && !strings.HasPrefix(hash, t.WorkC) {
		return protocol.NewProtocolError(protocol.MintPoWInvalidHash, "invalid workc")
	}

	return nil
}

// 判断是否可以mint
func (t *IERC20Tick) CanMint(want, minted decimal.Decimal) error {
	// 判断单笔挖取数量是否超标
	if want.GreaterThan(t.Limit) {
		return protocol.NewProtocolError(protocol.MintAmountExceedLimit, fmt.Sprintf("invalid amount. %s > limit", want))
	}

	// 判断当前地址的剩余可挖数量
	walletRemain := t.WalletLimit.Sub(minted)
	if want.GreaterThan(walletRemain) {
		return protocol.NewProtocolError(protocol.MintAmountExceedLimit, fmt.Sprintf("invalid amount. %s > wallet wallet_remain(%s)", want, walletRemain))
	}

	// 判断剩余发行量是否足够
	remain := t.MaxSupply.Sub(t.Supply)
	if want.GreaterThan(remain) {
		return protocol.NewProtocolError(protocol.MintAmountExceedLimit, fmt.Sprintf("invalid amount. %s > remain_supply(%s)", want, remain))
	}

	return nil
}

// mint
func (t *IERC20Tick) Mint(blockNumber uint64, amount decimal.Decimal) {
	t.Supply = t.Supply.Add(amount)
	t.LastUpdatedAtBlock = blockNumber
}

func (t *IERC20Tick) Marshal() ([]byte, error) {
	return json.Marshal(t)
}

func (t *IERC20Tick) Unmarshal(bytes []byte) error {
	return json.Unmarshal(bytes, t)
}
