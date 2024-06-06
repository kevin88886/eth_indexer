package staking

import (
	"github.com/kevin88886/eth_indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
)

type PoolTickDetail struct {
	Index         int             `json:"idx"`            //
	Tick          string          `json:"tick"`           // tick
	Ratio         decimal.Decimal `json:"ratio"`          // 当前tick对应的奖励比例
	Amount        decimal.Decimal `json:"amount"`         // 当前质押数量
	MaxAmount     decimal.Decimal `json:"max_amount"`     // 最大质押数量
	HistoryAmount decimal.Decimal `json:"history_amount"` // 历史质押数量, 带有限制区块的才需要计算
}

func (s *PoolTickDetail) Copy() *PoolTickDetail {
	return &PoolTickDetail{
		Index:         s.Index,
		Tick:          s.Tick,
		Ratio:         s.Ratio.Copy(),
		Amount:        s.Amount.Copy(),
		MaxAmount:     s.MaxAmount.Copy(),
		HistoryAmount: s.HistoryAmount.Copy(),
	}
}

type StakingPoolDetail struct {
	Name        string                     `json:"name"`                 // 名称
	Owner       string                     `json:"owner"`                // 奖励池的所有者
	Admins      []string                   `json:"admins,omitempty"`     // 普通管理员. 部署池子时指定的管理员
	StartBlock  uint64                     `json:"start_block"`          // 池子的部署区块
	StopBlock   uint64                     `json:"stop_block,omitempty"` // 停止奖励区块
	TickDetails map[string]*PoolTickDetail `json:"details,omitempty"`    //
}

type StakingPool struct {
	Pool             string            `json:"pool,omitempty"`             // 池子的地址(本质上是一个EOA地址)
	PoolSubID        uint64            `json:"poolSubID,omitempty"`        // 池子ID，不是数据库的ID
	Detail           StakingPoolDetail `json:"detail"`                     // config
	LastUpdatedBlock uint64            `json:"lastUpdatedBlock,omitempty"` // 最后更新时的区块号

	positions map[string]*StakingPosition // 仓位信息. map(staker => positionsByPoolID)
}

func NewStakingPool(command *protocol.ConfigStakeCommand) *StakingPool {

	var details = make(map[string]*PoolTickDetail)
	for idx, tickWithRatio := range command.Details {
		details[tickWithRatio.Tick] = &PoolTickDetail{
			Index:     idx,
			Tick:      tickWithRatio.Tick,
			Ratio:     tickWithRatio.RewardsRatioPerBlock,
			MaxAmount: tickWithRatio.MaxAmount,
			Amount:    decimal.Zero,
		}
	}

	return &StakingPool{
		Pool:      command.Pool,
		PoolSubID: command.PoolSubID,
		Detail: StakingPoolDetail{
			Name:        command.Name,
			Owner:       command.Owner,
			Admins:      command.Admins,
			StartBlock:  command.BlockNumber,
			StopBlock:   command.StopBlock,
			TickDetails: details,
		},
		LastUpdatedBlock: command.BlockNumber,
		positions:        make(map[string]*StakingPosition),
	}
}

func (p *StakingPool) getPosition(staker string) *StakingPosition {
	if p.positions == nil {
		return nil
	}

	// 查询质押人仓位
	position, existed := p.positions[staker]
	if !existed {
		return nil
	}

	return position
}

func (p *StakingPool) setPosition(position *StakingPosition) {
	if p.positions == nil {
		p.positions = make(map[string]*StakingPosition)
	}

	p.positions[position.Staker] = position
}

func (p *StakingPool) IsAmin(address string) bool {
	if address == p.Detail.Owner {
		return true
	}

	for _, admin := range p.Detail.Admins {
		if admin == address {
			return true
		}
	}

	return false
}

// 是否限期
func (p *StakingPool) IsTimeLimited() bool {
	return p.Detail.StopBlock != 0
}

func (p *StakingPool) IsEnd(currBlock uint64) bool {
	return p.Detail.StopBlock != 0 && p.Detail.StopBlock < currBlock
}

func (p *StakingPool) CanStaking(blockNumber uint64, tick string, amount decimal.Decimal) error {

	detail, existed := p.Detail.TickDetails[tick]
	if !existed {
		return protocol.NewProtocolError(protocol.StakingTickUnsupported, "tick unsupported")
	}

	if detail.Ratio.IsZero() {
		return protocol.NewProtocolError(protocol.StakingTickUnsupported, "tick unsupported")
	}

	// ====== 下面是限期池子的校验
	if p.IsTimeLimited() {
		// 判断池子是否到期
		if blockNumber >= p.Detail.StopBlock {
			return protocol.NewProtocolError(protocol.StakingPoolAlreadyStopped, "pool already stopped")
		}

		// 判断池子是否满了
		if detail.MaxAmount.Sub(detail.Amount).LessThan(amount) {
			return protocol.NewProtocolError(protocol.StakingPoolIsFulled, "pool is fulled")
		}
	}

	return nil
}

func (p *StakingPool) CanUnStaking(blockNumber uint64, tick string, amount decimal.Decimal) error {

	detail, existed := p.Detail.TickDetails[tick]
	if !existed {
		return protocol.NewProtocolError(protocol.StakingTickUnsupported, "tick unsupported")
	}

	// 池子中剩余的质押数量 < 要取消质押的数量
	if detail.Amount.LessThan(amount) {
		return protocol.NewProtocolError(protocol.UnStakingErrStakeAmountInsufficient, "invalid amount")
	}

	// ====== 下面是限期池子的校验
	if p.IsTimeLimited() {
		if blockNumber <= p.Detail.StopBlock {
			return protocol.NewProtocolError(protocol.UnStakingErrNotYetUnlocked, "not yet unlocked")
		}
	}

	return nil
}

func (p *StakingPool) CalcAvailableRewards(blockNumber uint64, staker string) decimal.Decimal {
	position := p.getPosition(staker)
	if position == nil {
		return decimal.Zero
	}

	if p.IsTimeLimited() {
		blockNumber = min(blockNumber, p.Detail.StopBlock)
	}

	return position.CalcAvailableRewards(blockNumber)
}

func (p *StakingPool) UpdatePool(command *protocol.ConfigStakeCommand) error {
	if p.IsTimeLimited() {
		return p.updateTimeLimitedPool(command)
	} else {
		return p.updateUnlimitedPool(command)
	}
}

// 更新无限制池
func (p *StakingPool) updateUnlimitedPool(command *protocol.ConfigStakeCommand) error {

	tickDetails := make(map[string]*PoolTickDetail)
	for _, info := range p.Detail.TickDetails {
		// 总质押数量为0的就可以移除了
		if info.Amount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		// 重置奖励比例
		info.Ratio = decimal.Zero
		tickDetails[info.Tick] = info
	}

	for idx, item := range command.Details {
		tick, existed := tickDetails[item.Tick]
		if !existed {
			tickDetails[item.Tick] = &PoolTickDetail{
				Index:  idx,
				Tick:   item.Tick,
				Ratio:  item.RewardsRatioPerBlock,
				Amount: decimal.Zero,
			}
		} else {
			tick.Index = idx
			tick.Ratio = item.RewardsRatioPerBlock
		}
	}

	// 重置所有仓位信息
	for _, position := range p.positions {

		// 结算奖励
		p.settleRewards(command.BlockNumber, position)

		// 重置每个区块的奖励
		position.ResetRewardsPerBlock(command.BlockNumber, tickDetails)
	}

	// 更新配置
	p.Detail.Name = command.Name
	p.Detail.TickDetails = tickDetails
	p.Detail.Admins = command.Admins
	p.LastUpdatedBlock = command.BlockNumber

	return nil
}

func (p *StakingPool) updateTimeLimitedPool(command *protocol.ConfigStakeCommand) error {

	if p.IsEnd(command.BlockNumber) {
		return protocol.NewProtocolError(protocol.StakingPoolIsEnded, "pool is ended")
	}

	tickDetails := make(map[string]*PoolTickDetail)
	for _, info := range p.Detail.TickDetails {
		// 总质押数量为0的就可以移除了
		if info.Amount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		// 重置奖励比例
		info.Ratio = decimal.Zero
		info.MaxAmount = decimal.Zero // TODO: z 最大数量是否可变更
		tickDetails[info.Tick] = info
	}

	for idx, item := range command.Details {
		tick, existed := tickDetails[item.Tick]
		if !existed {
			tickDetails[item.Tick] = &PoolTickDetail{
				Index:     idx,
				Tick:      item.Tick,
				Ratio:     item.RewardsRatioPerBlock,
				MaxAmount: item.MaxAmount,
				Amount:    decimal.Zero,
			}
		} else {
			// 如果最大数量小于当前质押数量, 则返回错误
			if item.MaxAmount.LessThan(tick.Amount) {
				return protocol.NewProtocolError(protocol.StakingPoolMaxAmountLessThanCurrentAmount, "max amount less than current amount")
			}

			tick.Index = idx
			tick.Ratio = item.RewardsRatioPerBlock
			tick.MaxAmount = item.MaxAmount
		}
	}

	// 重置所有仓位信息
	for _, position := range p.positions {

		// 结算奖励
		p.settleRewards(command.BlockNumber, position)

		// 重置每个区块的奖励
		position.ResetRewardsPerBlock(command.BlockNumber, tickDetails)
	}

	// 更新配置
	p.Detail.Name = command.Name
	p.Detail.TickDetails = tickDetails
	p.Detail.Admins = command.Admins
	p.LastUpdatedBlock = command.BlockNumber

	return nil
}

// 质押
func (p *StakingPool) Staking(blockNumber uint64, staker, tick string, amount decimal.Decimal) error {

	// 判断是否可以质押
	if err := p.CanStaking(blockNumber, tick, amount); err != nil {
		return err
	}

	// 查询质押的代币相关的奖励
	detail, _ := p.Detail.TickDetails[tick]

	// 查询质押人仓位
	position := p.getPosition(staker)
	if position == nil {
		position = NewStakingPosition(blockNumber, p.Pool, p.PoolSubID, staker) // 不存在则创建
		p.setPosition(position)
	} else {
		p.settleRewards(blockNumber, position)
	}

	// 质押, 质押不存在失败
	_ = position.Staking(blockNumber, detail.Tick, detail.Ratio, amount)

	// 更新质押池信息
	detail.Amount = detail.Amount.Add(amount) // 更新池子的总质押数量
	if p.IsTimeLimited() {
		detail.HistoryAmount = detail.HistoryAmount.Add(amount)
	}
	p.LastUpdatedBlock = blockNumber

	return nil
}

// 取消质押
func (p *StakingPool) UnStaking(blockNumber uint64, staker string, tick string, amount decimal.Decimal) error {

	// 判断是否可以取消质押
	if err := p.CanUnStaking(blockNumber, tick, amount); err != nil {
		return err
	}

	detail, _ := p.Detail.TickDetails[tick]

	// 查询质押人仓位
	position := p.getPosition(staker)
	if position == nil {
		return protocol.NewProtocolError(protocol.UnStakingErrNoStake, "no stake")
	}

	p.settleRewards(blockNumber, position) // 先结算仓位的奖励

	// 取消质押
	if err := position.UnStaking(blockNumber, detail.Tick, detail.Ratio, amount); err != nil {
		return err
	}

	detail.Amount = detail.Amount.Sub(amount) // 更新池子的总质押数量
	p.LastUpdatedBlock = blockNumber

	return nil
}

// 使用奖励, 返回实际使用数量
func (p *StakingPool) UseRewards(blockNumber uint64, staker string, amount decimal.Decimal) decimal.Decimal {

	position := p.getPosition(staker)
	if position == nil {
		return decimal.Zero
	}

	p.settleRewards(blockNumber, position)

	return position.UseRewards(blockNumber, amount)
}

func (p *StakingPool) settleRewards(blockNumber uint64, position *StakingPosition) {

	if p.IsTimeLimited() {
		blockNumber = min(blockNumber, p.Detail.StopBlock)
	}

	position.SettleRewards(blockNumber)
}
