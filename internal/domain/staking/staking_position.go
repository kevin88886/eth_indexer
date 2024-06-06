package staking

import (
	"time"

	"github.com/kevin88886/eth_indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
)

type PositionTickDetail struct {
	Tick   string          `json:"tick"`   // tick
	Ratio  decimal.Decimal `json:"ratio"`  // 当前tick对应的奖励比例
	Amount decimal.Decimal `json:"amount"` // 当前质押数量
}

type StakingPosition struct {
	PoolAddress      string
	PoolSubID        uint64
	Staker           string                         // 质押人
	TickDetails      map[string]*PositionTickDetail // 质押人的具体仓位信息。tick => amount
	RewardsPerBlock  decimal.Decimal                // 用户每个块的奖励
	Debt             decimal.Decimal                // 用户一共使用了多少的奖励
	AccReward        decimal.Decimal                // 用户累积了多少奖励
	LastRewardBlock  uint64                         // 用户上一次获得奖励的区块
	LastUpdatedBlock uint64                         // 最后更新的区块
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewStakingPosition(blockNumber uint64, pool string, poolID uint64, staker string) *StakingPosition {
	return &StakingPosition{
		PoolAddress:      pool,
		PoolSubID:        poolID,
		Staker:           staker,
		TickDetails:      make(map[string]*PositionTickDetail),
		RewardsPerBlock:  decimal.Zero,
		Debt:             decimal.Zero,
		AccReward:        decimal.Zero,
		LastRewardBlock:  blockNumber,
		LastUpdatedBlock: blockNumber,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
	}
}

// 计算剩余的可用奖励
func (s *StakingPosition) calculateRemainingAvailableRewards() decimal.Decimal {
	return s.AccReward.Sub(s.Debt)
}

// 计算未领取的奖励
func (s *StakingPosition) calculateUnclaimedRewards(blockNumber uint64) decimal.Decimal {
	if blockNumber <= s.LastRewardBlock {
		return decimal.Zero
	}

	return s.RewardsPerBlock.Mul(decimal.NewFromInt(int64(blockNumber - s.LastRewardBlock)))
}

func (s *StakingPosition) CalcAvailableRewards(blockNumber uint64) decimal.Decimal {
	remainingRewards := s.calculateRemainingAvailableRewards()
	unclaimedRewards := s.calculateUnclaimedRewards(blockNumber)
	return remainingRewards.Add(unclaimedRewards)
}

// 结算当前仓位的奖励点, 返回新增的奖励点数
func (s *StakingPosition) SettleRewards(blockNumber uint64) decimal.Decimal {
	if blockNumber <= s.LastRewardBlock {
		return decimal.Zero
	}

	unclaimedRewards := s.calculateUnclaimedRewards(blockNumber)
	s.AccReward = s.AccReward.Add(unclaimedRewards)
	s.LastRewardBlock = blockNumber
	return unclaimedRewards
}

// 重新计算当前仓位每个块的奖励
func (s *StakingPosition) ResetRewardsPerBlock(blockNumber uint64, ticks map[string]*PoolTickDetail) {

	// 直接重置所有奖励分配相关的参数
	var (
		rewardsPerBlock = decimal.Zero
		details         = make(map[string]*PositionTickDetail)
	)

	for _, detail := range s.TickDetails {
		// 忽略无效字段
		if detail.Amount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		detail.Ratio = decimal.Zero
		details[detail.Tick] = detail
	}

	for _, item := range ticks {
		// 查询质押数量, 更新配置
		detail, existed := details[item.Tick]
		if !existed {
			// 不存在, 新增一条记录
			detail = &PositionTickDetail{
				Tick:   item.Tick,
				Ratio:  item.Ratio,
				Amount: decimal.Zero,
			}
			details[item.Tick] = detail
		} else {
			// 已存在，更新奖励比例
			detail.Ratio = item.Ratio
		}

		// 如果奖励比例 <= 0 或者 质押数量 <= 0, 则不必计算, 跳过
		if detail.Ratio.LessThanOrEqual(decimal.Zero) || detail.Amount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		// 记录并统计
		rewardsPerBlock = rewardsPerBlock.Add(detail.Amount.Mul(detail.Ratio))
	}

	// 更新每个块的奖励
	s.TickDetails = details
	s.RewardsPerBlock = rewardsPerBlock
	s.LastUpdatedBlock = blockNumber
	return
}

// 使用奖励, 返回实际使用的奖励点数
func (s *StakingPosition) UseRewards(blockNumber uint64, useAmount decimal.Decimal) decimal.Decimal {

	if useAmount.IsZero() {
		return useAmount
	}

	if useAmount.LessThanOrEqual(decimal.Zero) {
		panic("logic error")
	}

	// 判断奖励数是否充足
	available := s.calculateRemainingAvailableRewards()

	// 计算实际使用数量
	realUseAmount := decimal.Min(useAmount, available)

	s.Debt = s.Debt.Add(realUseAmount)
	s.LastUpdatedBlock = blockNumber
	return realUseAmount
}

func (s *StakingPosition) Staking(blockNumber uint64, tick string, ratio, amount decimal.Decimal) error {

	detail, existed := s.TickDetails[tick]
	if !existed {
		detail = &PositionTickDetail{
			Tick:   tick,
			Ratio:  ratio,
			Amount: decimal.Zero,
		}
		s.TickDetails[detail.Tick] = detail
	}

	detail.Amount = detail.Amount.Add(amount)
	s.RewardsPerBlock = s.RewardsPerBlock.Add(amount.Mul(ratio))
	s.LastUpdatedBlock = blockNumber
	return nil
}

func (s *StakingPosition) UnStaking(blockNumber uint64, tick string, ratio, amount decimal.Decimal) error {

	detail, existed := s.TickDetails[tick]
	if !existed {
		return protocol.NewProtocolError(protocol.UnStakingErrNoStake, "no stake")
	}

	if amount.GreaterThan(detail.Amount) {
		return protocol.NewProtocolError(protocol.UnStakingErrStakeAmountInsufficient, "insufficient stake amount")
	}

	detail.Amount = detail.Amount.Sub(amount)
	s.RewardsPerBlock = s.RewardsPerBlock.Sub(amount.Mul(ratio))
	s.LastUpdatedBlock = blockNumber
	return nil
}
