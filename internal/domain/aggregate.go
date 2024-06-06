package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/kevin88886/eth_indexer/internal/domain/balance"
	"github.com/kevin88886/eth_indexer/internal/domain/protocol"
	"github.com/kevin88886/eth_indexer/internal/domain/staking"
	"github.com/kevin88886/eth_indexer/internal/domain/tick"
	"github.com/shopspring/decimal"
)

type AggregateRoot struct {
	// initialize state
	PreviousBlock uint64                                  // 上一个带有事件的区块号
	Block         *Block                                  // 区块号
	TicksMap      map[string]tick.Tick                    // 聚合根相关的ticks
	BalancesMap   map[balance.BalanceKey]*balance.Balance // 聚合根相关的balances
	Signatures    map[string]*IERC20TransferredEvent      // 聚合根相关的签名事件
	StakingPools  map[string]*staking.PoolAggregate       // 聚合根相关的质押池

	// config
	invalidTxHashMap map[string]struct{} // 无效交易Hash列表
	feeStartBlock    uint64              // 开始收手续费的区块

	// runtime
	mintFlag map[string]struct{}
	Events   []Event
}

func NewBlockAggregate(previous uint64, block *Block, invalidTxHashMap map[string]struct{}, feeStartBlock uint64) *AggregateRoot {
	if invalidTxHashMap == nil {
		invalidTxHashMap = make(map[string]struct{})
	}

	return &AggregateRoot{
		PreviousBlock:    previous,
		Block:            block,
		TicksMap:         make(map[string]tick.Tick),
		BalancesMap:      make(map[balance.BalanceKey]*balance.Balance),
		Signatures:       make(map[string]*IERC20TransferredEvent),
		StakingPools:     make(map[string]*staking.PoolAggregate),
		invalidTxHashMap: invalidTxHashMap,
		feeStartBlock:    feeStartBlock,
		mintFlag:         make(map[string]struct{}),
		Events:           nil,
	}
}

func (root *AggregateRoot) checkTxHash(txHash string) error {
	if _, existed := root.invalidTxHashMap[txHash]; existed {
		return protocol.NewProtocolError(protocol.InvalidTxHash, "invalid tx hash")
	}

	return nil
}

func (root *AggregateRoot) getOrCreateBalance(address, tick string) *balance.Balance {
	key := balance.NewBalanceKey(address, tick)
	entity, existed := root.BalancesMap[key]
	if existed {
		return entity
	}

	entity = &balance.Balance{
		ID:               0,
		Address:          address,
		Tick:             tick,
		Available:        decimal.Zero,
		Freeze:           decimal.Zero,
		MintedAmount:     decimal.Zero,
		LastUpdatedBlock: 0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	root.BalancesMap[key] = entity

	return entity
}

func (root *AggregateRoot) isMinted(address, tick string) bool {
	key := fmt.Sprintf("%s-%s", address, tick)
	_, existed := root.mintFlag[key]
	return existed
}

func (root *AggregateRoot) markMinted(address, tick string) {
	key := fmt.Sprintf("%s-%s", address, tick)
	root.mintFlag[key] = struct{}{}
}

func (root *AggregateRoot) resetMintFlag() {
	root.mintFlag = make(map[string]struct{})
}

type totalShare struct {
	PoWTotalShare decimal.Decimal
	PoSTotalShare decimal.Decimal
}

func (t *totalShare) AddPoSShare(share decimal.Decimal) {
	t.PoSTotalShare = t.PoSTotalShare.Add(share)
}

func (t *totalShare) AddPoWShare(share decimal.Decimal) {
	t.PoWTotalShare = t.PoWTotalShare.Add(share)
}

// 统计 pow mint 相关数据
func (root *AggregateRoot) calculatePoWMintShare() map[string]*totalShare {

	shares := make(map[string]*totalShare)

	for _, transaction := range root.Block.Transactions {
		// 跳过无效数据
		if transaction.IsProcessed || transaction.IERCTransaction == nil {
			continue
		}

		// 判断交易类型是否是 pow mint
		command, ok := transaction.IERCTransaction.(*protocol.MintPoWCommand)
		if !ok {
			continue
		}

		tickName := command.Tick()

		// 判断tick存不存在
		t, existed := root.TicksMap[tickName]
		if !existed {
			continue
		}

		// 判断tick类型、协议是否正确
		tickEntity, ok := t.(*tick.IERCPoWTick)
		if !ok || tickEntity.Protocol != command.Protocol {
			continue
		}

		// 判断是否存在积分对应的质押池
		pool, err := root.getPoolAggregate(tickEntity.Rule.PosPool)
		if err != nil {
			panic(err) // deploy 的时候已经判断了池子是否存在, 如果这里获取不到, 说明是程序逻辑出问题了
		}

		ts, existed := shares[tickName]
		if !existed {
			ts = new(totalShare)
			shares[tickName] = ts
		}

		// 判断是否已经挖过了
		if root.isMinted(command.From, tickName) {
			// 如果已经统计过份额了，直接标记为已处理, 避免有人通过漏洞超额mint
			transaction.Code = int32(protocol.MintErrTickMinted)
			transaction.Remark = "has been minted"
			transaction.IsProcessed = true
			transaction.UpdatedAt = time.Now()
			continue
		}

		var canMint bool

		switch {
		case command.IsDPoS() && command.IsPoW():

			// 判断 pow 份额是否为零
			share := root.calcPoWShare(command, tickEntity)
			if share.IsZero() {
				continue
			}

			points := command.Points()

			// 在指定区块后, 必须满足最小奖励点数
			if command.BlockNumber > protocol.DPoSMintMintPointsLimitBlockHeight &&
				points.LessThan(decimal.NewFromInt(protocol.DPoSMintMinPoints)) {
				points = decimal.Zero
			}

			// 判断 pos 份额是否足够
			if !pool.CanUseRewards(command.BlockNumber, command.From, points) {
				continue
			}

			ts.AddPoWShare(share) // 统计所有pow mint的份额
			ts.AddPoSShare(points)
			canMint = true

		case command.IsDPoS():
			// 判断是否有足够的奖励点
			points := command.Points()
			if command.BlockNumber > protocol.DPoSMintMintPointsLimitBlockHeight &&
				points.LessThan(decimal.NewFromInt(protocol.DPoSMintMinPoints)) {
				continue
			}

			if !pool.CanUseRewards(command.BlockNumber, command.From, points) {
				continue
			}

			ts.AddPoSShare(points) // 统计所有pos mint的奖励点
			canMint = true

		case command.IsPoW():
			// 判断份额是否为零
			share := root.calcPoWShare(command, tickEntity)
			if share.IsZero() {
				continue
			}

			ts.AddPoWShare(share) // 统计所有pow mint的份额
			canMint = true
		}

		if canMint {
			root.markMinted(command.From, tickName)
		}
	}

	root.resetMintFlag()

	return shares
}

func (root *AggregateRoot) calcPoWShare(tx *protocol.MintPoWCommand, tickEntity *tick.IERCPoWTick) decimal.Decimal {

	var diff = max(tx.Block(), tx.BlockNumber) - min(tx.Block(), tx.BlockNumber)
	if diff > 5 {
		return decimal.Zero
	}

	// 统计交易hash的总份额
	return tickEntity.CalculateMintShareBasedOnHash(tx.BlockNumber, tx.TxHash)
}

func (root *AggregateRoot) Handle() {
	fmt.Println("处理区快交易:")
	shares := root.calculatePoWMintShare()

	for _, transaction := range root.Block.Transactions {
		fmt.Println("transaction123:")
		fmt.Println(&transaction)
		fmt.Println(*transaction)
		if transaction.IsProcessed {
			continue
		}

		transaction.IsProcessed = true
		transaction.UpdatedAt = time.Now()

		if transaction.IERCTransaction == nil {
			continue
		}

		// 协议处理
		var err error
		switch tx := transaction.IERCTransaction.(type) {
		case *protocol.DeployCommand:
			err = root.HandleDeploy(tx)

		case *protocol.MintCommand:
			err = root.HandleMint(tx)

		case *protocol.DeployPoWCommand:
			err = root.handleDeployPow(tx)

		case *protocol.MintPoWCommand:
			powMintTotalShare, posMintTotalShare := decimal.Zero, decimal.Zero
			if ts, existed := shares[tx.Tick()]; existed {
				powMintTotalShare = ts.PoWTotalShare
				posMintTotalShare = ts.PoSTotalShare
			}
			err = root.handleMintPoW(tx, powMintTotalShare, posMintTotalShare)

		case *protocol.ModifyCommand:
			err = root.handleModify(tx)

		case *protocol.ClaimAirdropCommand:
			err = root.handleClaimAirdrop(tx)

		case *protocol.TransferCommand:
			err = root.HandleTransfer(tx)

		case *protocol.UnfreezeSellCommand:
			err = root.HandleUnfreezeSell(tx)

		case *protocol.FreezeSellCommandV4:
			err = root.HandleFreezeSellV4(tx)

		case *protocol.FreezeSellCommand:
			err = root.HandleFreezeSell(tx)

		case *protocol.ProxyTransferCommandV4:
			err = root.HandleProxyTransferV4(tx)

		case *protocol.ProxyTransferCommand:
			err = root.HandleProxyTransfer(tx)

		case *protocol.ConfigStakeCommand:
			err = root.handleConfigStaking(tx)

		case *protocol.StakingCommand:
			switch tx.Operate {
			case protocol.OpStaking:
				err = root.handleStaking(tx)
			case protocol.OpUnStaking:
				err = root.handleUnStaking(tx)
			case protocol.OpProxyUnStaking:
				err = root.handleProxyUnStaking(tx)
			}
		default:
			fmt.Println("无法处理的区块交易数据:")
			fmt.Println(tx.String())
		}

		if err != nil {
			var pErr *protocol.ProtocolError
			if errors.As(err, &pErr) {
				transaction.Code = pErr.Code()
				transaction.Remark = pErr.Message()
			} else {
				transaction.Code = int32(protocol.UnknownError)
				transaction.Remark = err.Error()
			}
			continue
		}
	}
}

// ==================== about tick: deploy & mint ====================

func (root *AggregateRoot) HandleDeploy(command *protocol.DeployCommand) (err error) {

	event := &IERC20TickCreatedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERC20TickCreated{
			Protocol:    command.Protocol,
			Operate:     command.Operate,
			Tick:        command.Tick,
			Decimals:    command.Decimals,
			MaxSupply:   command.MaxSupply,
			Limit:       command.MintLimitOfSingleTx,
			WalletLimit: command.MintLimitOfWallet,
			WorkC:       command.Workc,
			Nonce:       command.Nonce,
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}
	defer func() {
		event.SetError(err)
		root.Events = append(root.Events, event)
	}()

	// 检查tick是否已存在
	if _, existed := root.TicksMap[command.Tick]; existed {
		return protocol.NewProtocolError(protocol.TickExited, "tick already existed")
	}

	root.TicksMap[command.Tick] = tick.NewTickFromDeployCommand(command)

	return nil
}

func (root *AggregateRoot) HandleMint(command *protocol.MintCommand) (err error) {

	// 排除无效mint的hash
	if err = root.checkTxHash(command.TxHash); err != nil {
		return
	}

	ee := &IERC20MintedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERC20Minted{
			Protocol:     command.Protocol,
			Operate:      command.Operate,
			Tick:         command.Tick,
			From:         protocol.ZeroAddress, // mint 由 零地址 划转资金到 miner地址
			To:           command.From,
			MintedAmount: command.Amount,
			Gas:          command.Gas,
			GasPrice:     command.GasPrice,
			Nonce:        command.Nonce,
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}

	defer func() {
		ee.SetError(err)
		root.Events = append(root.Events, ee)
	}()

	// 加载tick
	tickEntity, existed := root.TicksMap[command.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not existed")
	}

	// 判断是否已经mint过了
	if root.isMinted(command.From, command.Tick) {
		return protocol.NewProtocolError(protocol.MintErrTickMinted, "has been minted")
	}

	ierc20TickEntity, ok := tickEntity.(*tick.IERC20Tick)
	if !ok {
		return protocol.NewProtocolError(protocol.MintErrTickProtocolNoMatch, "tick protocol no match")
	}

	// 验证 workc
	if err = ierc20TickEntity.ValidateHash(command.TxHash); err != nil {
		return err
	}

	// 验证mint数量是否符合要求
	minerBalance := root.getOrCreateBalance(command.From, command.Tick)
	if err = ierc20TickEntity.CanMint(command.Amount, minerBalance.MintedAmount); err != nil {
		return err
	}

	// 更新状态
	ierc20TickEntity.Mint(command.BlockNumber, command.Amount) // 更新 tick 的发行量
	minerBalance.AddMint(command.BlockNumber, command.Amount)  // 更新 miner 的资产信息
	root.markMinted(command.From, command.Tick)                // 标记 miner 在这个区块 mint 过了

	return
}

func (root *AggregateRoot) handleDeployPow(command *protocol.DeployPoWCommand) (err error) {

	event := &IERCPoWTickCreatedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERCPoWTickCreated{
			Protocol:   command.Protocol,
			Operate:    command.Operate,
			Tick:       command.Tick,
			Decimals:   command.Decimals,
			MaxSupply:  command.MaxSupply,
			Tokenomics: command.TokenomicsDetails,
			Rule:       command.DistributionRule,
			Creator:    command.From,
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}
	defer func() {
		event.SetError(err)
		root.Events = append(root.Events, event)
	}()

	// 检查tick是否已存在
	if _, existed := root.TicksMap[command.Tick]; existed {
		return protocol.NewProtocolError(protocol.TickExited, "tick already existed")
	}

	// 检查奖励池是否存在
	if _, err = root.getPoolAggregate(command.DistributionRule.PosPool); err != nil {
		return err
	}

	root.TicksMap[command.Tick] = tick.NewIERCPoWTickFromDeployCommand(command)

	return
}

func (root *AggregateRoot) handleMintPoW(command *protocol.MintPoWCommand, powTotalShare decimal.Decimal, posTotalShare decimal.Decimal) (err error) {

	ee := &IERCPoWMintedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERCPoWMinted{
			Protocol:        command.Protocol,
			Operate:         command.Operate,
			Tick:            command.Tick(),
			From:            protocol.ZeroAddress, // mint 由 零地址 划转资金到 miner地址
			To:              command.From,
			IsPoW:           command.IsPoW(),
			PoWTotalShare:   powTotalShare,
			PoWMinerShare:   decimal.Zero,
			PoWMintedAmount: decimal.Zero,
			IsPoS:           command.IsDPoS(),
			PoSTotalShare:   posTotalShare,
			PoSMinerShare:   decimal.Zero,
			PoSMintedAmount: decimal.Zero,
			PoSPointsSource: "",
			Gas:             command.Gas,
			GasPrice:        command.GasPrice,
			Nonce:           command.Nonce(),
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}
	root.Events = append(root.Events, ee) // 先加进事件列表, 后面可能存在burn事件

	defer func() {
		ee.SetError(err)
	}()

	tickName := command.Tick()
	// 判断是否已经mint过了
	if root.isMinted(command.From, tickName) {
		return protocol.NewProtocolError(protocol.MintErrTickMinted, "has been minted")
	}

	// 获取tick
	t, existed := root.TicksMap[tickName]
	if !existed {
		return protocol.NewProtocolError(protocol.MintErrTickNotFound, "tick not found")
	}

	// 判断是否支持pow
	tickEntity, ok := t.(*tick.IERCPoWTick)
	if !ok || tickEntity.Protocol != command.Protocol {
		return protocol.NewProtocolError(protocol.MintErrTickNotSupportPoW, "not support pow")
	}

	// 获取奖励池
	pool, err := root.getPoolAggregate(tickEntity.Rule.PosPool)
	if err != nil {
		return err
	}

	params := &tick.PoWMintParams{
		CurrentBlock:   command.BlockNumber,
		EffectiveBlock: command.Block(),
		IsPoW:          command.IsPoW(),
		IsDPoS:         command.IsDPoS(),
		TotalPoWShare:  powTotalShare,
		MinerPoWShare:  decimal.Zero,
		TotalPoSShare:  posTotalShare,
		MinerPoSShare:  decimal.Zero,
	}

	switch {
	case command.IsDPoS() && command.IsPoW():
		// 计算份额
		params.MinerPoWShare = tickEntity.CalculateMintShareBasedOnHash(command.BlockNumber, command.TxHash)
		if params.MinerPoWShare.IsZero() {
			return protocol.NewProtocolError(protocol.MintErrPoWShareZero, "invalid pow mint")
		}

		// 判断奖励点数是否足够
		points := command.Points()

		// 在指定区块后, 如果奖励点数小于最小值, 则不参与pos
		if command.BlockNumber > protocol.DPoSMintMintPointsLimitBlockHeight &&
			points.LessThan(decimal.NewFromInt(protocol.DPoSMintMinPoints)) {
			points = decimal.Zero
			params.IsDPoS = false
		}

		// 判断奖励是否足够
		if !pool.CanUseRewards(command.BlockNumber, command.From, points) {
			return protocol.NewProtocolError(protocol.UseRewardsErrRewardsInsufficient, "point insufficient")
		}

		params.MinerPoSShare = points

	case command.IsDPoS():
		points := command.Points()

		if command.BlockNumber > protocol.DPoSMintMintPointsLimitBlockHeight && points.LessThan(decimal.NewFromInt(protocol.DPoSMintMinPoints)) {
			return protocol.NewProtocolError(protocol.MintErrDPoSMintPointsTooLow, "point too low")
		}

		// 判断奖励是否足够
		if !pool.CanUseRewards(command.BlockNumber, command.From, points) {
			return protocol.NewProtocolError(protocol.UseRewardsErrRewardsInsufficient, "point insufficient")
		}

		params.MinerPoSShare = points

	case command.IsPoW():
		params.MinerPoWShare = tickEntity.CalculateMintShareBasedOnHash(command.BlockNumber, command.TxHash)
	}

	// 判断是否可以mint
	if err = tickEntity.CanMint(params); err != nil {
		return err
	}

	// TODO: z 临时处理，检查参数合法性，如果出现错误, 直接panic程序, 等待修复
	if err := params.Validate(); err != nil {
		log.Errorf("mint params error. command: %v, params: %v", command, params)
		panic(err)
	}

	// 如果是pos, 扣除奖励点
	if command.IsDPoS() && !params.MinerPoSShare.IsZero() {
		// 扣除奖励点
		if err = pool.UseRewards(command.BlockNumber, command.From, params.MinerPoSShare); err != nil {
			return err
		}
	}

	// mint
	powMintedAmount, posMintedAmount := tickEntity.Mint(params)

	// 更新 mint 余额
	minerBalance := root.getOrCreateBalance(command.From, tickName)
	minerBalance.AddMint(command.BlockNumber, powMintedAmount.Add(posMintedAmount))

	// 更新事件数据
	ee.Data.PoWMinerShare = params.MinerPoWShare
	ee.Data.PoWMintedAmount = powMintedAmount
	ee.Data.PoSMinerShare = command.Points()
	ee.Data.PoSPointsSource = pool.PoolAddress
	ee.Data.PoSMintedAmount = posMintedAmount

	// 尝试销毁多余的数量
	powBurnAmount, posBurnAmount := tickEntity.Burn()
	burnAmount := powBurnAmount.Add(posBurnAmount)
	// 判断是否有销毁的数量
	if burnAmount.GreaterThan(decimal.Zero) {
		burnEvent := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: 1, // burn 在 mint 之后, 所以填1
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol: command.Protocol,
				Operate:  command.Operate,
				Tick:     tickEntity.Tick,
				From:     protocol.ZeroAddress, // 销毁, 由零地址转到零地址
				To:       protocol.ZeroAddress,
				Amount:   burnAmount,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}

		// 黑洞地址
		blackHoleBalance := root.getOrCreateBalance(protocol.ZeroAddress, tickEntity.Tick)
		blackHoleBalance.AddAvailable(root.Block.Number, burnAmount)

		root.Events = append(root.Events, burnEvent)
	}

	root.markMinted(command.From, tickName)

	return
}

func (root *AggregateRoot) handleModify(command *protocol.ModifyCommand) (err error) {

	event := &IERCPoWTickCreatedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERCPoWTickCreated{
			Protocol:  command.Protocol,
			Operate:   command.Operate,
			Tick:      command.Tick,
			MaxSupply: command.MaxSupply,
			Creator:   command.From,
		},
		EventAt: command.EventAt,
	}
	root.Events = append(root.Events, event)
	defer func() {
		event.SetError(err)
	}()

	tickEntity, existed := root.TicksMap[command.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not existed")
	}

	powTickEntity, ok := tickEntity.(*tick.IERCPoWTick)
	if !ok {
		return protocol.NewProtocolError(protocol.ErrTickProtocolNoMatch, "tick protocol no match")
	}

	err = powTickEntity.UpdateMaxSupply(command.BlockNumber, command.From, command.MaxSupply)
	if err != nil {
		switch {
		case errors.Is(err, tick.ErrNoPermission):
			err = protocol.NewProtocolError(protocol.ErrUpdateMaxSupplyNoPermission, err.Error())

		case errors.Is(err, tick.ErrMaxAmountLessThanSupply):
			err = protocol.NewProtocolError(protocol.ErrUpdateAmountLessThanSupply, err.Error())

		default:
			err = protocol.NewProtocolError(protocol.ErrUpdateFailed, err.Error())
		}

		return err
	}

	return nil
}

func (root *AggregateRoot) handleClaimAirdrop(command *protocol.ClaimAirdropCommand) (err error) {

	ee := &IERCPoWMintedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERCPoWMinted{
			Protocol:      command.Protocol,
			Operate:       command.Operate,
			Tick:          command.Tick,
			From:          protocol.ZeroAddress, // mint 由 零地址 划转资金到 miner地址
			To:            command.From,
			IsAirdrop:     true,
			AirdropAmount: command.ClaimAmount,
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}
	root.Events = append(root.Events, ee)
	defer func() {
		ee.SetError(err)
	}()

	tickEntity, existed := root.TicksMap[command.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not existed")
	}

	powTickEntity, ok := tickEntity.(*tick.IERCPoWTick)
	if !ok {
		return protocol.NewProtocolError(protocol.MintErrTickProtocolNoMatch, "tick protocol no match")
	}

	err = powTickEntity.ClaimAirdrop(command.BlockNumber, command.From, command.ClaimAmount)
	if err != nil {
		switch {
		case errors.Is(err, tick.ErrNoPermission):
			err = protocol.NewProtocolError(protocol.MintErrNoPermissionToClaimAirdrop, err.Error())

		case errors.Is(err, tick.ErrAirdropAmountExceedsRemainSupply):
			err = protocol.NewProtocolError(protocol.MintErrAirdropAmountExceedsRemainSupply, err.Error())

		case errors.Is(err, tick.ErrInvalidAmount):
			err = protocol.NewProtocolError(protocol.MintErrInvalidAirdropAmount, err.Error())

		default:
			err = protocol.NewProtocolError(protocol.MintErrAirdropClaimFailed, err.Error())
		}

		return err
	}

	// 可用余额 < 划转数量, 划转失败
	fromBalance := root.getOrCreateBalance(command.From, command.Tick)
	fromBalance.AddAvailable(command.BlockNumber, command.ClaimAmount)

	return nil
}

// ==================== about tick: transfer ====================

func (root *AggregateRoot) HandleTransfer(command *protocol.TransferCommand) error {

	for idx, record := range command.Records {

		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol: record.Protocol,
				Operate:  record.Operate,
				Tick:     record.Tick,
				From:     record.From,
				To:       record.Recv,
				Amount:   record.Amount,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}

		root.Events = append(root.Events, ee)

		if err := root.checkTxHash(command.TxHash); err != nil {
			ee.SetError(err)
			continue
		}

		// 先校验Tick是否存在
		_, existed := root.TicksMap[record.Tick]
		if !existed {
			err := protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
			ee.SetError(err)
			continue
		}

		// 处理转账
		err := root.handleTransferRecord(record)
		if err != nil {
			ee.SetError(err)
			continue
		}
	}

	return nil
}

func (root *AggregateRoot) handleTransferRecord(record *protocol.TransferRecord) error {
	// 可用余额 < 划转数量, 划转失败
	fromBalance := root.getOrCreateBalance(record.From, record.Tick)

	if fromBalance.Available.LessThan(record.Amount) {
		return protocol.NewProtocolError(
			protocol.InsufficientAvailableFunds,
			fmt.Sprintf("insufficient balance. available(%s) < transfer(%s)", fromBalance.Available, record.Amount),
		)
	}

	toBalance := root.getOrCreateBalance(record.Recv, record.Tick)
	// 资金操作
	fromBalance.SubAvailable(root.Block.Number, record.Amount)
	toBalance.AddAvailable(root.Block.Number, record.Amount)
	return nil
}

// ==================== about trade: freeze & unfreeze & proxy_transfer ====================

func (root *AggregateRoot) HandleFreezeSell(command *protocol.FreezeSellCommand) error {
	fmt.Println("V3签名")
	fmt.Println(command)
	if err := root.checkTxHash(command.TxHash); err != nil {
		return err
	}

	buyerRemainEthValue := command.TxValue.Shift(-18).Sub(protocol.ServiceGasPrice.Shift(-18)) // 买方用于购买Tick的以太币数量
	for idx, record := range command.Records {

		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol:    record.Protocol,
				Operate:     record.Operate,
				Tick:        record.Tick,
				From:        record.Seller, // freezeSell 的 from、to 都是卖家地址, 表示从可用划转到冻结
				To:          record.Seller,
				Amount:      record.Amount,
				EthValue:    record.Value,
				GasPrice:    record.GasPrice,
				Nonce:       "",
				SignerNonce: record.SignNonce,
				Sign:        record.SellerSign,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}

		// 处理转账
		err := root.handleFreezeRecord(&record, buyerRemainEthValue)

		if err != nil {
			ee.SetError(err)
		} else {
			value := record.Value
			if root.Block.Number > root.feeStartBlock {
				value = value.Mul(protocol.ServiceFee) // TODO: z 服务费不能设置为零
			}
			buyerRemainEthValue = buyerRemainEthValue.Sub(value)
			// 如果冻结成功了, 更新签名使用情况
			root.Signatures[record.SellerSign] = ee
		}

		// 记录变更信息
		root.Events = append(root.Events, ee)
	}

	return nil
}

func (root *AggregateRoot) HandleFreezeSellV4(command *protocol.FreezeSellCommandV4) error {
	fmt.Println("V4签名:FreezeSell")
	fmt.Println(command)
	if err := root.checkTxHash(command.TxHash); err != nil {
		return err
	}

	buyerRemainEthValue := command.TxValue.Shift(-18).Sub(protocol.ServiceGasPrice.Shift(-18)) // 买方用于购买Tick的以太币数量
	for idx, record := range command.Records {

		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol:    record.Protocol,
				Operate:     record.Operate,
				Tick:        record.Tick,
				From:        record.Seller, // freezeSell 的 from、to 都是卖家地址, 表示从可用划转到冻结
				To:          record.Seller,
				Amount:      record.Amount,
				EthValue:    record.Value,
				Nonce:       "",
				SignerNonce: record.SignNonce,
				Sign:        record.SellerSign,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}

		// 处理转账
		err := root.handleFreezeRecordV4(&record, buyerRemainEthValue)

		if err != nil {
			ee.SetError(err)
		} else {
			value := record.Value
			if root.Block.Number > root.feeStartBlock {
				value = value.Mul(protocol.ServiceFee) // TODO: z 服务费不能设置为零
			}
			buyerRemainEthValue = buyerRemainEthValue.Sub(value)
			// 如果冻结成功了, 更新签名使用情况
			root.Signatures[record.SellerSign] = ee
		}

		// 记录变更信息
		root.Events = append(root.Events, ee)
	}

	return nil
}

func (root *AggregateRoot) HandleFreezeSellBundle(command *protocol.FreezeSellCommand) error {
	return nil
}

func (root *AggregateRoot) handleFreezeRecord(record *protocol.FreezeRecord, buyerRemainEthValue decimal.Decimal) error {

	// 先校验Tick是否存在
	_, existed := root.TicksMap[record.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	// 参数检查
	if err := record.ValidateParams(); err != nil {
		return err
	}

	// 签名验证
	if err := record.ValidateSignature(); err != nil {
		return err
	}

	// 检查签名是否已被使用
	if ee, existed := root.Signatures[record.SellerSign]; existed {
		switch ee.Data.Operate {
		case protocol.OpFreezeSell:
			return protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used. freeze_sell")
		case protocol.OpProxyTransfer:
			return protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used. proxy_transfer")

		// 签名被解冻了, 可以冻结
		case protocol.OpUnfreezeSell:

		// 正常逻辑不可能进入default分支
		default:
			panic("FreezeSell Sign Error")
		}
	}

	// 验证买方交易中携带的以太币数量
	// 额外扣取服务费

	// TODO: z 在指定区块后开始校验服务费
	value := record.Value
	if root.Block.Number > root.feeStartBlock {
		value = value.Mul(protocol.ServiceFee) // TODO: z
	}
	if buyerRemainEthValue.LessThan(value) {
		return protocol.NewProtocolError(
			protocol.InsufficientValue,
			fmt.Sprintf("insufficient value. remainEthValue(%s) < sellerValue(%s)", buyerRemainEthValue, record.Value),
		)
	}

	// 验证卖方的Tick可用余额
	// 可用余额 必须大于 要冻结的数量
	sellerBalance := root.getOrCreateBalance(record.Seller, record.Tick)
	if sellerBalance.Available.LessThan(record.Amount) {
		return protocol.NewProtocolError(
			protocol.InsufficientAvailableFunds,
			fmt.Sprintf("insufficient balance. avaliable(%v) < wantFreeze(%v)", sellerBalance.Available, record.Amount),
		)
	}

	// 资金操作
	sellerBalance.FreezeBalance(root.Block.Number, record.Amount)

	return nil
}

func (root *AggregateRoot) handleFreezeRecordV4(record *protocol.FreezeRecordV4, buyerRemainEthValue decimal.Decimal) error {

	// 先校验Tick是否存在
	_, existed := root.TicksMap[record.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	// 参数检查
	if err := record.ValidateParamsV4(); err != nil {
		return err
	}

	// 签名验证
	if err := record.ValidateSignatureV4(); err != nil {
		return err
	}

	// 检查签名是否已被使用
	if ee, existed := root.Signatures[record.SellerSign]; existed {
		switch ee.Data.Operate {
		case protocol.OpFreezeSell:
			return protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used. freeze_sell")
		case protocol.OpProxyTransfer:
			return protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used. proxy_transfer")

		// 签名被解冻了, 可以冻结
		case protocol.OpUnfreezeSell:

		// 正常逻辑不可能进入default分支
		default:
			panic("FreezeSell Sign Error")
		}
	}

	// 验证买方交易中携带的以太币数量
	// 额外扣取服务费

	// TODO: z 在指定区块后开始校验服务费
	value := record.Value
	if root.Block.Number > root.feeStartBlock {
		value = value.Mul(protocol.ServiceFee) // TODO: z
	}
	if buyerRemainEthValue.LessThan(value) {
		return protocol.NewProtocolError(
			protocol.InsufficientValue,
			fmt.Sprintf("insufficient value. remainEthValue(%s) < sellerValue(%s)", buyerRemainEthValue, record.Value),
		)
	}

	// 验证卖方的Tick可用余额
	// 可用余额 必须大于 要冻结的数量
	sellerBalance := root.getOrCreateBalance(record.Seller, record.Tick)
	if sellerBalance.Available.LessThan(record.Amount) {
		return protocol.NewProtocolError(
			protocol.InsufficientAvailableFunds,
			fmt.Sprintf("insufficient balance. avaliable(%v) < wantFreeze(%v)", sellerBalance.Available, record.Amount),
		)
	}

	// 资金操作
	sellerBalance.FreezeBalance(root.Block.Number, record.Amount)

	return nil
}

func (root *AggregateRoot) HandleUnfreezeSell(command *protocol.UnfreezeSellCommand) error {

	for idx, record := range command.Records {

		event, err := root.handleUnfreezeRecord(command.To, &record)

		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data:              event,
			ErrCode:           0,
			ErrReason:         "",
			EventAt:           command.EventAt,
		}
		if err != nil {
			ee.SetError(err)
		} else {
			root.Signatures[record.Sign] = ee
		}

		root.Events = append(root.Events, ee)
	}

	return nil
}

func (root *AggregateRoot) handleUnfreezeRecord(unfreezeAddress string, record *protocol.UnfreezeRecord) (*IERC20Transferred, *protocol.ProtocolError) {

	// TODO: z 缺少验证签名

	var event = &IERC20Transferred{
		Operate: protocol.OpUnfreezeSell,
	}

	// 检查签名是否已被使用
	ee, existed := root.Signatures[record.Sign]
	if !existed {
		return event, protocol.NewProtocolError(protocol.SignatureNotExist, "signature not exist")
	}

	tickEntity, existed := root.TicksMap[ee.Data.Tick]
	if !existed {
		return event, protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	// 根据 FreezeSell 数据重新填充数据
	event = &IERC20Transferred{
		Protocol:    tickEntity.GetProtocol(),
		Operate:     protocol.OpUnfreezeSell,
		Tick:        tickEntity.GetName(),
		From:        ee.To, // freezeSell 的 from、to 都是卖家地址, 表示从可用划转到冻结, 所以这里的 from、to 都是同一个人
		To:          ee.From,
		Amount:      ee.Data.Amount,
		EthValue:    ee.Data.EthValue,
		GasPrice:    ee.Data.GasPrice,
		Nonce:       ee.Data.Nonce,
		SignerNonce: ee.Data.SignerNonce,
		Sign:        ee.Data.Sign,
	}

	switch ee.Data.Operate {
	// 签名被冻结, 正常处理
	case protocol.OpFreezeSell:
		if ee.From != unfreezeAddress {
			return event, protocol.NewProtocolError(protocol.SignatureNotMatch, "signature address not match")
		}
		if strings.ToLower(ee.TxHash) != strings.ToLower(record.TxHash) {
			return event, protocol.NewProtocolError(protocol.SignatureNotMatch, "freeze hash not match")
		}

	// 签名已成交
	case protocol.OpProxyTransfer:
		return event, protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used. proxy_transfer")

	// 签名已经被解冻使用了
	case protocol.OpUnfreezeSell:
		return event, protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used. unfreeze_sell")

	// 正常逻辑不可能进入default分支
	default:
		panic("signature status error")
	}

	var unfreezeAmount = ee.Data.Amount

	sellerBalance := root.getOrCreateBalance(ee.Data.From, ee.Data.Tick)
	if sellerBalance.Freeze.LessThan(unfreezeAmount) {
		return event, protocol.NewProtocolError(
			protocol.InsufficientFreezeFunds,
			fmt.Sprintf("insufficient freeze funds. freeze(%v) < unfreeze(%v)", sellerBalance.Freeze, unfreezeAmount),
		)
	}

	// 资金操作
	sellerBalance.UnfreezeBalance(root.Block.Number, unfreezeAmount)

	return event, nil
}

func (root *AggregateRoot) HandleProxyTransfer(command *protocol.ProxyTransferCommand) error {
	fmt.Println("V3签名:ProxyTransfer")
	fmt.Println(command)
	if err := root.checkTxHash(command.TxHash); err != nil {
		return err
	}

	buyerRemainEthValue := command.TxValue.Shift(-18) // 买方用于购买Tick的以太币数量

	for idx, record := range command.Records {

		// 处理转账
		event, err := root.handleProxyTransferRecord(&record, buyerRemainEthValue)
		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data:              event,
			ErrCode:           0,
			ErrReason:         "",
			EventAt:           command.EventAt,
		}
		if err != nil {
			ee.SetError(err)
		} else {
			buyerRemainEthValue = buyerRemainEthValue.Sub(record.Value)
			root.Signatures[record.Sign] = ee
		}

		root.Events = append(root.Events, ee)
	}

	return nil
}

func (root *AggregateRoot) HandleProxyTransferV4(command *protocol.ProxyTransferCommandV4) error {
	fmt.Println("V4签名:ProxyTransfer")
	fmt.Println(command)
	if err := root.checkTxHash(command.TxHash); err != nil {
		return err
	}

	buyerRemainEthValue := command.TxValue.Shift(-18) // 买方用于购买Tick的以太币数量

	for idx, record := range command.Records {

		// 处理转账
		event, err := root.handleProxyTransferRecordV4(&record, buyerRemainEthValue)
		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data:              event,
			ErrCode:           0,
			ErrReason:         "",
			EventAt:           command.EventAt,
		}
		if err != nil {
			ee.SetError(err)
		} else {
			buyerRemainEthValue = buyerRemainEthValue.Sub(record.Value)
			root.Signatures[record.Sign] = ee
		}

		root.Events = append(root.Events, ee)
	}

	return nil
}

func (root *AggregateRoot) handleProxyTransferRecord(record *protocol.ProxyTransferRecord, buyerRemainEthValue decimal.Decimal) (*IERC20Transferred, error) {

	var event = &IERC20Transferred{
		Protocol:    record.Protocol,
		Operate:     record.Operate,
		Tick:        record.Tick,
		From:        record.From,
		To:          record.To,
		Amount:      record.Amount,
		EthValue:    record.Value,
		GasPrice:    decimal.Zero,
		Nonce:       "",
		SignerNonce: record.SignerNonce,
		Sign:        record.Sign,
	}

	tickEntity, existed := root.TicksMap[record.Tick]
	if !existed {
		return event, protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	// 填充 event 数据
	event.Protocol = tickEntity.GetProtocol()

	// 验证参数
	if err := record.ValidateParams(); err != nil {
		return event, err.(*protocol.ProtocolError)
	}

	// 签名校验
	if err := record.ValidateSignature(); err != nil {
		return event, err.(*protocol.ProtocolError)
	}

	// 检查签名是否已被使用
	ee, existed := root.Signatures[record.Sign]
	if !existed {
		return event, protocol.NewProtocolError(protocol.SignatureNotExist, "freeze sell not exist")
	}

	switch ee.Data.Operate {
	case protocol.OpProxyTransfer:
		return event, protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used")
	case protocol.OpUnfreezeSell:
		return event, protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already unfreeze")
	// 正常冻结
	case protocol.OpFreezeSell:

	default:
		panic("proxy transfer signature error")
	}

	// 验证交易中携带的以太坊数量是否足够
	if root.Block.Number > root.feeStartBlock && buyerRemainEthValue.LessThan(record.Value.Mul(protocol.HandlingFeeAmount)) {
		return event, protocol.NewProtocolError(
			protocol.InsufficientValue,
			fmt.Sprintf("insufficient value. remainETHValue(%s) < recordValue(%s)", buyerRemainEthValue, record.Value),
		)
	}

	// 检查冻结余额
	fromBalance := root.getOrCreateBalance(record.From, record.Tick)
	if fromBalance.Freeze.LessThan(record.Amount) {
		return event, protocol.NewProtocolError(
			protocol.InsufficientFreezeFunds,
			fmt.Sprintf("from insufficient balance. freeze(%s) < transfer(%s)", fromBalance.Freeze, record.Amount),
		)
	}

	toBalance := root.getOrCreateBalance(record.To, record.Tick)

	// 资金划转
	fromBalance.SubFreeze(root.Block.Number, record.Amount)
	toBalance.AddAvailable(root.Block.Number, record.Amount)

	return event, nil
}

func (root *AggregateRoot) handleProxyTransferRecordV4(record *protocol.ProxyTransferRecordV4, buyerRemainEthValue decimal.Decimal) (*IERC20Transferred, error) {

	var event = &IERC20Transferred{
		Protocol:    record.Protocol,
		Operate:     record.Operate,
		Tick:        record.Tick,
		From:        record.From,
		To:          record.To,
		Amount:      record.Amount,
		EthValue:    record.Value,
		GasPrice:    decimal.Zero,
		Nonce:       "",
		SignerNonce: record.SignerNonce,
		Sign:        record.Sign,
	}

	tickEntity, existed := root.TicksMap[record.Tick]
	if !existed {
		return event, protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	// 填充 event 数据
	event.Protocol = tickEntity.GetProtocol()

	// 验证参数
	if err := record.ValidateParams(); err != nil {
		return event, err.(*protocol.ProtocolError)
	}

	// 签名校验
	if err := record.ValidateSignature(); err != nil {
		return event, err.(*protocol.ProtocolError)
	}

	// 检查签名是否已被使用
	ee, existed := root.Signatures[record.Sign]
	if !existed {
		return event, protocol.NewProtocolError(protocol.SignatureNotExist, "freeze sell not exist")
	}

	switch ee.Data.Operate {
	case protocol.OpProxyTransfer:
		return event, protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used")
	case protocol.OpUnfreezeSell:
		return event, protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already unfreeze")
	// 正常冻结
	case protocol.OpFreezeSell:

	default:
		panic("proxy transfer signature error")
	}

	// 验证交易中携带的以太坊数量是否足够
	if root.Block.Number > root.feeStartBlock && buyerRemainEthValue.LessThan(record.Value.Mul(protocol.HandlingFeeAmount)) {
		return event, protocol.NewProtocolError(
			protocol.InsufficientValue,
			fmt.Sprintf("insufficient value. remainETHValue(%s) < recordValue(%s)", buyerRemainEthValue, record.Value),
		)
	}

	// 检查冻结余额
	fromBalance := root.getOrCreateBalance(record.From, record.Tick)
	if fromBalance.Freeze.LessThan(record.Amount) {
		return event, protocol.NewProtocolError(
			protocol.InsufficientFreezeFunds,
			fmt.Sprintf("from insufficient balance. freeze(%s) < transfer(%s)", fromBalance.Freeze, record.Amount),
		)
	}

	toBalance := root.getOrCreateBalance(record.To, record.Tick)

	// 资金划转
	fromBalance.SubFreeze(root.Block.Number, record.Amount)
	toBalance.AddAvailable(root.Block.Number, record.Amount)

	return event, nil
}

// ==================== about staking: config pool & stake & unstake & proxy_unstake ====================

func (root *AggregateRoot) getPoolAggregate(pool string) (*staking.PoolAggregate, error) {
	poolRoot, existed := root.StakingPools[pool]
	if !existed {
		// 在预处理的时候会尝试创建池子聚合, 所以这里不可能查不到
		return nil, protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	return poolRoot, nil
}

func (root *AggregateRoot) handleConfigStaking(command *protocol.ConfigStakeCommand) (err error) {

	ee := &StakingPoolUpdatedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &StakingPoolUpdated{
			Protocol:  command.Protocol,
			Operate:   command.Operate,
			From:      command.From,
			To:        command.To,
			Pool:      command.Pool,
			PoolID:    command.PoolSubID,
			Name:      command.Name,
			Owner:     command.Owner,
			Admins:    command.Admins,
			Details:   command.Details,
			StopBlock: command.StopBlock,
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}

	defer func() {
		ee.SetError(err)
		root.Events = append(root.Events, ee)
	}()

	// 校验tick是否存在
	for _, record := range command.Details {
		if _, existed := root.TicksMap[record.Tick]; !existed {
			return protocol.NewProtocolError(protocol.StakingTickNotExisted, "invalid tick")
		}
	}

	poolRoot, err := root.getPoolAggregate(command.Pool)
	if err != nil {
		poolRoot = staking.NewPoolAggregate(command.Pool, command.Owner)
		root.StakingPools[poolRoot.PoolAddress] = poolRoot
		err = nil
	}

	return poolRoot.UpdatePool(command)
}

func (root *AggregateRoot) handleStaking(command *protocol.StakingCommand) error {

	poolRoot, err := root.getPoolAggregate(command.Pool)
	if err != nil {
		return err
	}

	if !poolRoot.SubPoolIsExisted(command.PoolSubID) {
		return protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	for idx, record := range command.Details {

		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol: record.Protocol,
				Operate:  record.Operate,
				Tick:     record.Tick,
				From:     record.Staker, // staker
				To:       record.Pool,   // pool
				Amount:   record.Amount,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}
		root.Events = append(root.Events, ee)

		if err := root.handleStakingRecord(poolRoot, record); err != nil {
			ee.SetError(err)
		}
	}

	return nil
}

func (root *AggregateRoot) handleStakingRecord(pool *staking.PoolAggregate, record *protocol.StakingDetail) error {

	_, existed := root.TicksMap[record.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	stakerBalance := root.getOrCreateBalance(record.Staker, record.Tick)

	// 判断 staker 可用资金是否充足
	if record.Amount.GreaterThan(stakerBalance.Available) {
		return protocol.NewProtocolError(
			protocol.InsufficientAvailableFunds,
			fmt.Sprintf("insufficient balance. available(%s) < stake(%s)", stakerBalance.Available, record.Amount),
		)
	}

	// 质押
	err := pool.Staking(root.Block.Number, record.PoolSubID, record.Staker, record.Tick, record.Amount)
	if err != nil {
		return err
	}

	// 减去可用
	stakerBalance.SubAvailable(root.Block.Number, record.Amount)
	// 划转至目标地址的冻结上
	poolBalance := root.getOrCreateBalance(pool.PoolAddress, record.Tick)
	poolBalance.AddFreeze(root.Block.Number, record.Amount)

	return nil
}

func (root *AggregateRoot) handleUnStaking(command *protocol.StakingCommand) error {

	poolRoot, err := root.getPoolAggregate(command.Pool)
	if err != nil {
		return err
	}

	if !poolRoot.SubPoolIsExisted(command.PoolSubID) {
		return protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	for idx, record := range command.Details {
		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol: record.Protocol,
				Operate:  record.Operate,
				Tick:     record.Tick,
				From:     record.Pool,   // pool
				To:       record.Staker, // staker
				Amount:   record.Amount,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}
		root.Events = append(root.Events, ee)

		if err := root.handleUnStakingRecord(poolRoot, record); err != nil {
			ee.SetError(err)
		}
	}

	return nil
}

func (root *AggregateRoot) handleUnStakingRecord(pool *staking.PoolAggregate, record *protocol.StakingDetail) error {

	_, existed := root.TicksMap[record.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	// 进行取消质押操作
	if err := pool.UnStaking(root.Block.Number, record.PoolSubID, record.Staker, record.Tick, record.Amount); err != nil {
		return err
	}

	// 检查 pool 的冻结资产, 理论上来说能取消质押说明资金是足够的
	poolBalance := root.getOrCreateBalance(pool.PoolAddress, record.Tick)
	if poolBalance.Freeze.LessThan(record.Amount) {
		// 进这里说明数据出错了
		panic("pool freeze funds error, data error")
	}

	poolBalance.SubFreeze(root.Block.Number, record.Amount)

	// 添加可用
	stakerBalance := root.getOrCreateBalance(record.Staker, record.Tick)
	stakerBalance.AddAvailable(root.Block.Number, record.Amount)

	return nil
}

func (root *AggregateRoot) handleProxyUnStaking(command *protocol.StakingCommand) error {

	// 判断池子是否存在
	poolRoot, err := root.getPoolAggregate(command.Pool)
	if err != nil {
		return err
	}

	if !poolRoot.SubPoolIsExisted(command.PoolSubID) {
		return protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	// 判断是否是管理员
	if !poolRoot.IsAdmin(command.PoolSubID, command.From) {
		return protocol.NewProtocolError(protocol.ProxyUnStakingErrNotAdmin, "not admin")
	}

	for idx, record := range command.Details {
		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol: record.Protocol,
				Operate:  record.Operate,
				Tick:     record.Tick,
				From:     record.Pool,   // pool
				To:       record.Staker, // staker
				Amount:   record.Amount,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}
		root.Events = append(root.Events, ee)

		err := root.handleUnStakingRecord(poolRoot, record)
		if err != nil {
			ee.SetError(err)
		}
	}

	return nil
}
