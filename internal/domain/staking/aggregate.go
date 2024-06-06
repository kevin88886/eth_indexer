package staking

import (
	"sort"

	"github.com/kevin88886/eth_indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
)

type PoolAggregate struct {
	PoolAddress string                  // 池子主地址
	Owner       string                  // 池子所有者
	pools       map[uint64]*StakingPool // 子池信息
	//positions   map[string]map[uint64]*StakingPosition // 仓位信息. map(staker => positionsByPoolID)
}

func NewPoolAggregate(pool string, owner string) *PoolAggregate {
	return &PoolAggregate{
		PoolAddress: pool,
		Owner:       owner,
		pools:       make(map[uint64]*StakingPool),
		//positions:   make(map[string]map[uint64]*StakingPosition),
	}
}

func (p *PoolAggregate) InitPool(pool *StakingPool) {
	if pool.Pool != p.PoolAddress {
		return
	}

	p.pools[pool.PoolSubID] = pool
}

func (p *PoolAggregate) InitPosition(position *StakingPosition) {
	pool, existed := p.pools[position.PoolSubID]
	if !existed {
		return
	}

	pool.setPosition(position)
}

func (p *PoolAggregate) IsAdmin(poolSubID uint64, address string) bool {
	if address == p.Owner {
		return true
	}

	pool, existed := p.pools[poolSubID]
	if !existed {
		return false
	}

	return pool.IsAmin(address)
}

// 获取质押池
func (p *PoolAggregate) GetStakingPools() []*StakingPool {

	var result = make([]*StakingPool, 0, len(p.pools))
	for _, pool := range p.pools {
		result = append(result, pool)
	}

	// 根据PoolID对返回的结果进行排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].PoolSubID > result[j].PoolSubID
	})

	return result
}

func (p *PoolAggregate) SubPoolIsExisted(poolSubID uint64) bool {
	_, existed := p.pools[poolSubID]
	return existed
}

// 获取所有仓位
func (p *PoolAggregate) GetStakingPositions() []*StakingPosition {
	var positions []*StakingPosition
	for _, pool := range p.pools {
		for _, position := range pool.positions {
			positions = append(positions, position)
		}
	}

	return positions
}

// 更新池子配置
func (p *PoolAggregate) UpdatePool(command *protocol.ConfigStakeCommand) error {

	if command.Owner != p.Owner {
		return protocol.NewProtocolError(protocol.StakeConfigNoPermission, "no permission")
	}

	if command.Pool != p.PoolAddress {
		return protocol.NewProtocolError(protocol.StakeConfigPoolNotMatch, "not match")
	}

	if pool, existed := p.pools[command.PoolSubID]; existed {
		return pool.UpdatePool(command)
	} else {
		pool = NewStakingPool(command)
		p.pools[pool.PoolSubID] = pool
	}

	return nil
}

// 质押
func (p *PoolAggregate) Staking(blockNumber uint64, poolID uint64, staker, tick string, amount decimal.Decimal) error {

	pool, existed := p.pools[poolID]
	if !existed {
		return protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	return pool.Staking(blockNumber, staker, tick, amount)
}

// 取消质押
func (p *PoolAggregate) UnStaking(blockNumber uint64, poolID uint64, staker string, tick string, amount decimal.Decimal) error {
	// 获取对应的质押池
	pool, existed := p.pools[poolID]
	if !existed {
		return protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	return pool.UnStaking(blockNumber, staker, tick, amount)
}

// 判断奖励是否足够使用
func (p *PoolAggregate) CanUseRewards(blockNumber uint64, staker string, amount decimal.Decimal) bool {

	var rewards = decimal.Zero
	for _, pool := range p.pools {
		rewards = rewards.Add(pool.CalcAvailableRewards(blockNumber, staker))
	}

	// 剩余的可用奖励 >= 要使用的奖励点数
	return rewards.GreaterThanOrEqual(amount)
}

// 使用建议
func (p *PoolAggregate) UseRewards(blockNumber uint64, staker string, amount decimal.Decimal) error {

	useRewards := amount
	for _, pool := range p.pools {

		realUseAmount := pool.UseRewards(blockNumber, staker, useRewards)

		useRewards = useRewards.Sub(realUseAmount)
		if useRewards.IsZero() {
			return nil
		}
	}

	panic("logic error") // TODO: z 别问为什么，图快
}
