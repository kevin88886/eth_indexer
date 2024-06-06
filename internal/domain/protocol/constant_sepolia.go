//go:build sepolia
// +build sepolia

package protocol

const (
	PlatformAddress = "0x1878d3363a02f1b5e13ce15287c5c29515000656"

	// 限制双挖的区块高度
	DPoSMintMintPointsLimitBlockHeight uint64 = 0
	// 禁止双挖的区块高度
	DPoSDisableDualMiningBlockHeight uint64 = 5152670
	// 限制PoW挖矿的区块高度
	PoWMintLimitBlockHeight uint64 = 5182950
	// pos mint 最小点数
	DPoSMintMinPoints int64 = 1000
)
