package service

import (
	"context"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/kevin88886/eth_indexer/internal/conf"
	"github.com/kevin88886/eth_indexer/internal/domain"
	"github.com/kevin88886/eth_indexer/internal/domain/balance"
	"github.com/kevin88886/eth_indexer/internal/domain/protocol"
	"github.com/kevin88886/eth_indexer/internal/domain/staking"
	"github.com/kevin88886/eth_indexer/internal/domain/tick"
	"golang.org/x/sync/errgroup"
)

// 区块处理器, 负责处理区块数据
type BlockService struct {
	logger          *log.Helper
	blockRepo       domain.BlockRepository
	eventRepo       domain.EventRepository
	transactionRepo domain.TransactionRepository
	tickRepo        tick.TickRepository
	balanceRepo     balance.BalanceRepository
	stakingRepo     staking.StakingRepository

	// config
	invalidHashMap map[string]struct{} // 无效交易Hash. 来自配置文件
	feeStartBlock  uint64              // 开始收费的区块

	// runtime
	lastHandleBlock uint64 // 最后处理的区块, 只记录带有事件的区块
}

func NewBlockService(
	c *conf.Config,
	logger log.Logger,
	blockRepo domain.BlockRepository,
	eventRepo domain.EventRepository,
	transactionRepo domain.TransactionRepository,
	tickRepo tick.TickRepository,
	balanceRepo balance.BalanceRepository,
	stakingRepo staking.StakingRepository,
) (*BlockService, error) {
	lastBlock, err := eventRepo.GetBlockNumberByLastEvent(context.Background())
	if err != nil {
		return nil, err
	}

	return &BlockService{
		logger:          log.NewHelper(log.With(logger, "module", "BlockService")),
		blockRepo:       blockRepo,
		eventRepo:       eventRepo,
		transactionRepo: transactionRepo,
		tickRepo:        tickRepo,
		balanceRepo:     balanceRepo,
		stakingRepo:     stakingRepo,
		invalidHashMap:  c.InvalidTxHash,
		feeStartBlock:   c.Runtime.GetFeeStartBlock(),
		lastHandleBlock: lastBlock,
	}, nil
}

// 获取最后处理的区块
func (b *BlockService) GetLastHandleBlock() uint64 {
	return b.lastHandleBlock
}

// 同步区块
func (b *BlockService) SyncBlock(ctx context.Context, blocks []*domain.Block) error {
	return b.blockRepo.BulkSaveBlock(ctx, blocks)
}

// 处理区块
func (b *BlockService) HandleBlock(ctx context.Context, block *domain.Block) error {
	b.logger.Infof("start handle block. block_number: %d, transaction: %d", block.Number, len(block.Transactions))
	var (
		start      = time.Now()
		eventCount int
	)
	defer func() {
		b.logger.Infof("handle block done. block_number: %d, events: %d, duration: %v", block.Number, eventCount, time.Since(start))
	}()

	// 预处理，加载相关的所有数据
	aggregate, err := b.preprocessing(ctx, block)
	if err != nil {
		return err
	}

	// 处理区块中的交易
	aggregate.Handle()

	// 保存到数据库
	// TODO: z 失败重试
	if err := b.saveToDBWithTx(ctx, aggregate); err != nil {
		return err
	}

	eventCount = len(aggregate.Events)
	if len(aggregate.Events) != 0 {
		b.lastHandleBlock = aggregate.Block.Number // 更新最后处理区块
	}

	return nil
}

// 区块预处理
func (b *BlockService) preprocessing(ctx context.Context, block *domain.Block) (*domain.AggregateRoot, error) {

	var (
		aggregate = domain.NewBlockAggregate(b.lastHandleBlock, block, b.invalidHashMap, b.feeStartBlock)

		tickSet         = mapset.NewSet[string]()             // 记录当前区块涉及到的所有tick
		balanceSet      = mapset.NewSet[balance.BalanceKey]() // 记录当前区块涉及到的所有余额信息
		signatureSet    = mapset.NewSet[string]()             // 记录当前区块涉及到的所有签名
		unfreezeSignSet = mapset.NewSet[string]()             // 记录当前区块涉及到的解冻事件相关的签名
	)

	// 直接加载所有质押池
	pools, err := b.stakingRepo.LoadAllPools(ctx)
	if err != nil {
		return nil, err
	}

	aggregate.StakingPools = pools

loop:
	for _, transaction := range block.Transactions {

		// 有交易, 但是加载的时候解析失败了, 所以存在空的情况
		if transaction.IERCTransaction == nil {
			//b.logger.Debugf("ignore transactions that are not IERC20. tx: %v", transaction)
			transaction.IsProcessed = true
			continue loop
		}

		// 验证协议
		if err := transaction.IERCTransaction.Validate(); err != nil {
			transaction.Code = err.(*protocol.ProtocolError).Code()
			transaction.Remark = err.Error()
			transaction.IsProcessed = true
			//b.logger.Debugf("transaction validate failed. tx: %v error: %s", transaction, err)
			continue loop
		}

		b.logger.Debugf("preprocessing transaction: %v", transaction.IERCTransaction)

		// 协议处理
		switch t := transaction.IERCTransaction.(type) {
		// 部署
		case *protocol.DeployCommand:
			tickSet.Add(t.Tick)

		case *protocol.DeployPoWCommand:
			tickSet.Add(t.Tick)

		// 挖矿
		case *protocol.MintCommand:
			tickSet.Add(t.Tick)
			balanceSet.Add(balance.NewBalanceKey(t.From, t.Tick))

		case *protocol.MintPoWCommand:
			tickName := t.Tick()
			tickSet.Add(tickName)
			balanceSet.Add(balance.NewBalanceKey(t.From, tickName))
			// pow mint 可能存在销毁, 所以加载零地址的余额
			balanceSet.Add(balance.NewBalanceKey(protocol.ZeroAddress, tickName))

		case *protocol.ModifyCommand:
			tickSet.Add(t.Tick)

		case *protocol.ClaimAirdropCommand:
			tickSet.Add(t.Tick)
			balanceSet.Add(balance.NewBalanceKey(t.From, t.Tick))

		// 划转
		case *protocol.TransferCommand:
			for _, record := range t.Records {
				tickSet.Add(record.Tick)
				balanceSet.Add(balance.NewBalanceKey(record.From, record.Tick))
				balanceSet.Add(balance.NewBalanceKey(record.Recv, record.Tick))
			}

		// 冻结
		case *protocol.FreezeSellCommand:
			for _, record := range t.Records {
				tickSet.Add(record.Tick)
				balanceSet.Add(balance.NewBalanceKey(record.Seller, record.Tick))
				signatureSet.Add(record.SellerSign)
			}

		// 冻结V4
		case *protocol.FreezeSellCommandV4:
			for _, record := range t.Records {
				tickSet.Add(record.Tick)
				balanceSet.Add(balance.NewBalanceKey(record.Seller, record.Tick))
				signatureSet.Add(record.SellerSign)
			}

		// 解冻
		case *protocol.UnfreezeSellCommand:
			for _, record := range t.Records {
				signatureSet.Add(record.Sign)
				unfreezeSignSet.Add(record.Sign)
			}

		// 结算
		case *protocol.ProxyTransferCommand:
			for _, record := range t.Records {
				tickSet.Add(record.Tick)
				balanceSet.Add(balance.NewBalanceKey(record.From, record.Tick))
				balanceSet.Add(balance.NewBalanceKey(record.To, record.Tick))
				signatureSet.Add(record.Sign)
			}

		// 结算v4
		case *protocol.ProxyTransferCommandV4:
			for _, record := range t.Records {
				tickSet.Add(record.Tick)
				balanceSet.Add(balance.NewBalanceKey(record.From, record.Tick))
				balanceSet.Add(balance.NewBalanceKey(record.To, record.Tick))
				signatureSet.Add(record.Sign)
			}

		// 配置质押池
		case *protocol.ConfigStakeCommand:
			for _, record := range t.Details {
				tickSet.Add(record.Tick)
			}

		// 质押
		case *protocol.StakingCommand:
			for _, record := range t.Details {
				tickSet.Add(record.Tick)
				balanceSet.Add(balance.NewBalanceKey(record.Pool, record.Tick))
				balanceSet.Add(balance.NewBalanceKey(record.Staker, record.Tick))
			}

		default:
			transaction.Code = int32(protocol.InvalidProtocolParams)
			transaction.Remark = "invalid operate"
			transaction.IsProcessed = true
			continue loop
		}
	}

	//startAt := time.Now()
	//b.logger.Debugf("start load data. %s", startAt)
	//defer func() {
	//	b.logger.Debugf("load done. duration: %s", time.Since(startAt))
	//}()

	eg, gCtx := errgroup.WithContext(ctx)
	// 加载 tick 信息
	eg.Go(func() error {
		return b.loadTicks(gCtx, aggregate, tickSet.ToSlice())
	})

	// 加载持仓信息
	eg.Go(func() error {
		return b.loadBalances(gCtx, aggregate, balanceSet.ToSlice())
	})

	// 加载签名相关的成功事件, 对于一个签名，只加载最后成功的那个事件
	eg.Go(func() error {
		return b.loadEventsBySignature(gCtx, aggregate, signatureSet.ToSlice())
	})

	// 等待并发执行完成
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// 补全 unfreeze 事件相关的数据, 必须在上面这几个并发查询之后
	if err := b.loadUnfreezeEventRelatedData(ctx, aggregate, unfreezeSignSet); err != nil {
		return nil, err
	}

	return aggregate, nil
}

// 加载Tick
func (b *BlockService) loadTicks(ctx context.Context, root *domain.AggregateRoot, names []string) error {
	// 先检查是否已存在
	var queryTicks = make([]string, 0, len(names))
	for _, name := range names {
		_, existed := root.TicksMap[name]
		if existed {
			continue
		}

		queryTicks = append(queryTicks, name)
	}

	if len(queryTicks) == 0 {
		return nil
	}

	for _, tickName := range queryTicks {
		entity, err := b.tickRepo.Load(ctx, tickName)
		if err != nil {
			return err // database error
		}

		if entity == nil {
			continue
		}

		root.TicksMap[entity.GetName()] = entity
	}

	return nil
}

// 加载持仓信息
func (b *BlockService) loadBalances(ctx context.Context, root *domain.AggregateRoot, keys []balance.BalanceKey) error {

	var queries = make([]balance.BalanceKey, 0, len(keys))
	for _, key := range keys {
		_, existed := root.BalancesMap[key]
		if existed {
			continue
		}

		queries = append(queries, key)
	}

	if len(queries) == 0 {
		return nil
	}

	for _, key := range queries {
		entity, err := b.balanceRepo.Load(ctx, key)
		if err != nil {
			return err
		}

		if entity == nil {
			continue
		}

		root.BalancesMap[entity.Key()] = entity
	}

	return nil
}

// 根据签名加载事件
func (b *BlockService) loadEventsBySignature(ctx context.Context, root *domain.AggregateRoot, signs []string) error {
	signatures, err := b.eventRepo.QueryEventBySignature(ctx, signs)
	if err != nil {
		return err
	}

	for _, event := range signatures {
		if event == nil {
			continue
		}

		if e, ok := event.(*domain.IERC20TransferredEvent); ok && e.Data.Sign != "" {
			root.Signatures[e.Data.Sign] = e
		}
	}

	return nil
}

// 加载解冻事件相关的数据
func (b *BlockService) loadUnfreezeEventRelatedData(ctx context.Context, root *domain.AggregateRoot, unfreezeSignSet mapset.Set[string]) error {

	var (
		tickSet    = mapset.NewSet[string]()
		balanceSet = mapset.NewSet[balance.BalanceKey]()
	)

	for sign, e := range root.Signatures {

		if !unfreezeSignSet.Contains(sign) {
			continue
		}

		// 如果是 UnfreezeSell, 并且存在 FreezeSell 的签名, 还需要额外取查询对应的Tick、地址余额等信息
		if e.Data.Operate != protocol.OpFreezeSell {
			continue
		}

		_, existed := root.TicksMap[e.Data.Tick]
		if !existed {
			tickSet.Add(e.Data.Tick)
		}

		key := balance.NewBalanceKey(e.Data.From, e.Data.Tick)
		_, existed = root.BalancesMap[key]
		if !existed {
			balanceSet.Add(key)
		}
	}

	eg, gCtx := errgroup.WithContext(ctx)
	// 加载 tick 信息
	if tickSet.Cardinality() > 0 {
		eg.Go(func() error {
			return b.loadTicks(gCtx, root, tickSet.ToSlice())
		})
	}

	// 加载持仓信息
	if balanceSet.Cardinality() > 0 {
		eg.Go(func() error {
			return b.loadBalances(gCtx, root, balanceSet.ToSlice())
		})
	}

	return eg.Wait()
}

func (b *BlockService) saveToDBWithTx(ctx context.Context, root *domain.AggregateRoot) error {

	var (
		needUpdateTicks    = make([]tick.Tick, 0, len(root.TicksMap))
		needUpdateBalances = make([]*balance.Balance, 0, len(root.BalancesMap))
		pools              = poolsMapToSlice(root.StakingPools)
	)

	// 统计需要更新的 tick
	for _, entity := range root.TicksMap {
		if entity.LastUpdatedBlock() < root.Block.Number {
			continue
		}

		needUpdateTicks = append(needUpdateTicks, entity)
	}

	// 统计需要更新的 balance
	for _, entity := range root.BalancesMap {
		if entity.LastUpdatedBlock < root.Block.Number {
			continue
		}

		needUpdateBalances = append(needUpdateBalances, entity)
	}

	// 开启一个事务进行持久化保存
	err := b.transactionRepo.TransactionSave(ctx, func(ctxWithTx context.Context) error {
		// 更新区块信息
		if err := b.blockRepo.Update(ctxWithTx, root.Block); err != nil {
			return err
		}

		// 更新事件
		event := &domain.EventsByBlock{BlockNumber: root.Block.Number, Events: root.Events}
		if err := b.eventRepo.Save(ctxWithTx, event); err != nil {
			return err
		}

		// 更新tick
		if err := b.tickRepo.Save(ctxWithTx, needUpdateTicks...); err != nil {
			return err
		}

		// 更新balances
		if err := b.balanceRepo.Save(ctxWithTx, needUpdateBalances...); err != nil {
			return err
		}

		// 更新质押池信息
		if err := b.stakingRepo.Save(ctxWithTx, root.Block.Number, pools...); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 更新数据缓存
	return b.transactionRepo.UpdateCache(ctx, func(ctxWithUpdateKind context.Context) error {
		_ = b.tickRepo.Save(ctxWithUpdateKind, needUpdateTicks...)             // 更新tick缓存
		_ = b.balanceRepo.Save(ctxWithUpdateKind, needUpdateBalances...)       // 更新balance缓存
		_ = b.stakingRepo.Save(ctxWithUpdateKind, root.Block.Number, pools...) // 更新质押池缓存
		return nil
	})
}

func poolsMapToSlice(poolsMap map[string]*staking.PoolAggregate) []*staking.PoolAggregate {
	var result = make([]*staking.PoolAggregate, 0, len(poolsMap))
	for _, root := range poolsMap {
		result = append(result, root)
	}

	return result
}
