package protocol

import (
	"fmt"
)

type ProtocolErrCode int

const (
	ProtocolErr            ProtocolErrCode = iota + 0x0100 // 无效数据
	NotProtocolData                                        // 不是协议数据
	InvalidProtocolFormat                                  // 无效协议格式
	InvalidProtocolParams                                  // 无效协议参数
	UnknownProtocol                                        // 未知协议
	UnknownProtocolOperate                                 // 未知协议操作
	InvalidTxHash                                          // 无效交易hash, 为兼容老系统

	TickNotExist // Tick 不存在
	TickExited   // Tick 已存在

	InsufficientAvailableFunds // 可用资金不足
	InsufficientFreezeFunds    // 冻结资金不足
	InsufficientValue          // ETH value 不足

	SignatureNotExist    // 签名不存在
	SignatureAlreadyUsed // 签名已经使用
	SignatureNotMatch    // 签名不匹配

	MintErr                                 ProtocolErrCode = iota + 0x0200
	MintErrTickNotFound                                     // tick 不存在
	MintErrTickNotSupportPoW                                // tick 不支持 pow
	MintErrTickProtocolNoMatch                              // tick 的协议不匹配
	MintErrTickMinted                                       // tick 已经 mint 了
	MintPoWInvalidHash                                      // 无效hash
	MintPoSInvalidShare                                     // 无效份额
	MintAlreadyMinted                                       // mint完成
	MintAmountExceedLimit                                   // 数量超出限制
	MintInvalidBlock                                        // 无效区块
	MintBlockExpires                                        // 区块过期
	MintErrMaxAmountLessThanSupply                          // 最大数量小于总量
	MintErrNoPermissionToClaimAirdrop                       // 没有权限领取空投
	MintErrInvalidAirdropAmount                             // 无效空投数量
	MintErrAirdropAmountExceedsRemainSupply                 // 空投数量超过剩余数量
	MintErrAirdropClaimFailed                               // 空投领取失败

	InvalidSignature // 无效签名

	TickErr                        ProtocolErrCode = iota + 0x0300
	ErrTickProtocolNoMatch                         // tick 的协议不匹配
	ErrUpdateMaxSupplyNoPermission                 // 没有权限更新最大发行量
	ErrUpdateAmountLessThanSupply                  // 更新数量小于总量
	ErrUpdateFailed                                // 更新失败

	UnknownError ProtocolErrCode = iota + 0x0800

	StakingError                              ProtocolErrCode = iota + 0x0900
	StakingTickUnsupported                                    // Tick不支持
	StakingTickNotExisted                                     // Tick不存在
	StakingPoolNotFound                                       // 质押. 质押池不存在
	StakingPoolAlreadyStopped                                 // 质押. 质押池已关闭
	StakingPoolIsFulled                                       // 质押. 质押池已满
	StakingPoolIsEnded                                        // 质押. 质押池已结束
	StakingPoolMaxAmountLessThanCurrentAmount                 // 质押. 质押池最大数量小于当前数量
	StakeConfigPoolNotMatch                                   // 更新质押池. 池子不匹配
	StakeConfigNoPermission                                   // 更新质押池. 权限不足
	UnStakingErrNoStake                                       // 取消质押. 没有质押
	UnStakingErrStakeAmountInsufficient                       // 取消质押. 质押数量不够解冻
	UnStakingErrNotYetUnlocked                                // 取消质押. 还没到解锁时间
	ProxyUnStakingErrNotAdmin                                 // 取消质押. 不是管理员
	UseRewardsErrNoStake                                      // 使用奖励. 没有质押
	UseRewardsErrRewardsInsufficient                          // 使用奖励. 奖励点不足
	MintErrDPoSMintPointsTooLow                               // mint. 奖励点数太低
	MintErrPoWShareZero                                       // mint. pow份额为0
)

type ProtocolError struct {
	code    ProtocolErrCode
	message string
}

func (e *ProtocolError) Code() int32 { return int32(e.code) }

func (e *ProtocolError) Message() string { return e.message }

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("error code: %d, message: %s", e.code, e.message)
}

func NewProtocolError(code ProtocolErrCode, message string) *ProtocolError {
	return &ProtocolError{
		code:    code,
		message: message,
	}
}
