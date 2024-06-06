//go:build !sepolia

package protocol

const (
	PlatformAddress = "0x33302dbff493ed81ba2e7e35e2e8e833db023333"

	// 限制双挖的区块高度
	DPoSMintMintPointsLimitBlockHeight uint64 = 19033750
	// 禁止双挖的区块高度
	DPoSDisableDualMiningBlockHeight uint64 = 19085665
	// 限制PoW挖矿的区块高度
	PoWMintLimitBlockHeight uint64 = 19119100
	// pos mint 最小点数
	DPoSMintMinPoints int64 = 1000
)
