package protocol

import (
	"regexp"
	"strings"

	"github.com/kevin88886/eth_indexer/pkg/utils"
	"github.com/shopspring/decimal"
)

// 协议定义的常量
const (
	ZeroAddress = "0x0000000000000000000000000000000000000000"

	ProtocolHeader = `data:application/json,` // 协议头

	TickETHI = "ethi"

	SignatureTitle            = "ierc-20 one approve"
	CreateOrderSignatureTitle = "ierc-20 seller approve"

	// tick 名称长度最大限制
	TickMaxLength = 64
)

var (
	TickMaxSupplyLimit = decimal.RequireFromString("9999999999999999999999999999999")
	//ServiceFee         = decimal.RequireFromString("1.02") // 冻结服务费
	ServiceFee        = decimal.RequireFromString("1") // 冻结服务费
	WorkCRegexp       = regexp.MustCompile(`^0x[0-9a-f]{1,64}$`)
	ServiceGasPrice   = decimal.RequireFromString("0.0009") // 固定Gas费
	HandlingFeeAmount = decimal.RequireFromString("0.98")   // 扣除服务费的到账比例
)

func init() {
	// 零地址、平台地址检查
	if !utils.IsHexAddressWith0xPrefix(ZeroAddress) ||
		!utils.IsHexAddressWith0xPrefix(PlatformAddress) ||
		strings.ToLower(ZeroAddress) != ZeroAddress ||
		strings.ToLower(PlatformAddress) != PlatformAddress {
		panic("constant check error")
	}
}

type Protocol string

const (
	ProtocolTERC20  Protocol = "terc-20"  // 老版本协议, 只有 ethi 这个tick才支持
	ProtocolIERC20  Protocol = "ierc-20"  // ierc-20 协议
	ProtocolIERCPoW Protocol = "ierc-pow" // ierc-pow 协议
)

type Operate string

const (
	OpDeploy         = "deploy"
	OpMint           = "mint"
	OpTransfer       = "transfer"
	OpFreezeSell     = "freeze_sell"
	OpUnfreezeSell   = "unfreeze_sell"
	OpRefund         = "refund"
	OpProxyTransfer  = "proxy_transfer"
	OpStakeConfig    = "stake_config"
	OpStaking        = "stake"
	OpUnStaking      = "unstake"
	OpProxyUnStaking = "proxy_unstake"

	OpPoWModify       = "modify"
	OpPoWClaimAirdrop = "airdrop_claim"
)
