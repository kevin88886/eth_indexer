package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/kevin88886/eth_indexer/internal/conf"
	"github.com/kevin88886/eth_indexer/internal/domain"
	"github.com/kevin88886/eth_indexer/pkg/utils"
	"golang.org/x/sync/errgroup"
)

type IndexDomainService struct {
	ctx    context.Context
	cancel context.CancelFunc
	eg     *errgroup.Group

	fetcher   domain.BlockFetcher
	blockRepo domain.BlockRepository
	handler   *BlockService

	enableSync     bool   // 是否开启同步
	syncStartBlock uint64 // 同步开始起始区块号
	syncThreadsNum uint64 // 同步线程数量

	enableHandle   bool               // 是否开启处理
	handleEndBlock uint64             // 处理结束区块
	handleQueue    chan *domain.Block // 处理队列

	invalidHashMap map[string]struct{} // 无效交易Hash. 来自配置文件
	feeStartBlock  uint64
	status         *domain.BlockHandleStatus

	log log.Logger
}

func NewIndexApplication(
	data *conf.Config,
	log log.Logger,
	fetcher domain.BlockFetcher,
	blockRepo domain.BlockRepository,
	handler *BlockService,
) *IndexDomainService {

	ctx, cancel := context.WithCancel(context.Background())
	eg, gCtx := errgroup.WithContext(ctx)

	return &IndexDomainService{
		ctx:            gCtx,
		cancel:         cancel,
		eg:             eg,
		fetcher:        fetcher,
		blockRepo:      blockRepo,
		handler:        handler,
		enableSync:     data.Runtime.EnableSync,
		syncStartBlock: data.Runtime.SyncStartBlock,
		syncThreadsNum: max(data.Runtime.SyncThreadsNum, 1),
		enableHandle:   data.Runtime.EnableHandle,
		handleEndBlock: data.Runtime.HandleEndBlock,
		handleQueue:    make(chan *domain.Block, data.Runtime.HandleQueueSize),
		invalidHashMap: data.InvalidTxHash,
		feeStartBlock:  data.Runtime.GetFeeStartBlock(),
		status:         new(domain.BlockHandleStatus),
		log:            log,
	}
}

func (srv *IndexDomainService) Start(_ context.Context) error {
	helper := log.NewHelper(srv.log)
	helper.Info("start indexer service")
	defer helper.Info("quit indexer service...")

	if srv.eg == nil {
		return nil
	}

	// 初始化状态
	if err := srv.initStatus(); err != nil {
		return err
	}

	//  start sync
	if srv.enableSync {
		srv.eg.Go(utils.WithRetryCount(5, time.Second*15, time.Minute*3, srv.syncBlockLoop))
	}

	// start transaction handle
	if srv.enableHandle {
		srv.eg.Go(srv.loadBlockLoop)
		srv.eg.Go(srv.handleBlockLoop)
	}

	return srv.eg.Wait()
}

func (srv *IndexDomainService) Stop(_ context.Context) error {

	if srv.cancel != nil {
		srv.cancel()
	}

	if srv.eg != nil {
		return srv.eg.Wait()
	}

	return nil
}

func (srv *IndexDomainService) Status() *domain.BlockHandleStatus {
	return srv.status
}

func (srv *IndexDomainService) initStatus() error {

	status := &domain.BlockHandleStatus{
		LatestBlock:      nil,
		LastIndexedBlock: nil,
		LastSyncBlock:    nil,
	}

	eg, gCtx := errgroup.WithContext(srv.ctx)

	// 从节点获取最新高度
	eg.Go(func() error {
		latestBlock, err := srv.fetcher.GetBlockHeaderByNumber(gCtx, 0)
		if err != nil {
			return err
		}

		status.LatestBlock = latestBlock
		return nil
	})

	// 获取数据库最新索引到的区块
	eg.Go(func() error {
		indexed, err := srv.blockRepo.GetLastIndexedBlock(gCtx)
		if err != nil {
			return err
		}

		if indexed != nil {
			status.LastIndexedBlock = indexed
			return nil
		}

		indexed, err = srv.fetcher.GetBlockHeaderByNumber(gCtx, srv.syncStartBlock)
		if err != nil {
			return err
		}

		status.LastIndexedBlock = indexed
		return nil
	})

	eg.Go(func() error {
		syncBlock, err := srv.blockRepo.GetLastHandleBlock(gCtx)
		if err != nil {
			return err
		}

		status.LastSyncBlock = syncBlock
		return nil
	})

	srv.status = status

	return eg.Wait()
}

func (srv *IndexDomainService) syncBlockLoop() error {
	helper := log.NewHelper(log.With(srv.log, "method", "SyncBlockLoop"))
	helper.Info("start sync block loop")
	defer helper.Info("quit sync block loop")

	for {
		select {
		case <-srv.ctx.Done():
			return nil
		default:
		}

		helper.Infof("block handle status: %s", srv.status)

		status := srv.status

		switch {
		case status.LastIndexedBlock.Number+1 < status.LatestBlock.Number:

			var (
				indexStartNumber = status.LastIndexedBlock.Number + 1                                  // 索引开始区块
				indexEndNumber   = min(status.LatestBlock.Number, indexStartNumber+srv.syncThreadsNum) // 索引结束区块
				size             = indexEndNumber - indexStartNumber                                   // 大小
			)

			helper.Infof("fetch block. start_height: %d, end_height: %d, size: %d", indexStartNumber, indexEndNumber, size)
			blocks, err := srv.fetchBlocks(srv.ctx, indexStartNumber, size)
			if err != nil {
				helper.Errorf("fetch blocks failed. err: %s", err)
				return err
			}

			if len(blocks) == 0 {
				continue
			}

			lastIndexedBlock := status.LastIndexedBlock

			for _, block := range blocks {
				if lastIndexedBlock != nil && block.ParentHash != lastIndexedBlock.Hash {
					helper.Errorf("block data error. lastIndexedHash: %s, parentHash: %s", lastIndexedBlock.Hash, block.ParentHash)
					return errors.New("block rollback")
				}

				lastIndexedBlock = block.Header()
			}

			// 保存
			if err = srv.blockRepo.BulkSaveBlock(srv.ctx, blocks); err != nil {
				return err
			}

			// 更新最新索引区块
			status.LastIndexedBlock = lastIndexedBlock

		// 最新索引的区块等于最新的区块号, 更新区块号
		default:

			latestBlock, err := srv.fetcher.GetBlockHeaderByNumber(srv.ctx, 0)
			if err != nil {
				return err
			}

			switch {
			case latestBlock.Number == status.LatestBlock.Number:
				helper.Info("There is no latest block, please update after 10 seconds.")
				time.Sleep(time.Second * 10)

			case latestBlock.Number > status.LatestBlock.Number:
				helper.Infof("fetch latest block number. latest_block_number: %s", latestBlock)
				status.LatestBlock = latestBlock

			default:
				return fmt.Errorf("block data error, node: %d, local: %d", latestBlock.Number, status.LatestBlock.Number)
			}
		}
	}
}

func (srv *IndexDomainService) fetchBlocks(ctx context.Context, startAt uint64, size uint64) ([]*domain.Block, error) {

	var (
		gg, gCtx   = errgroup.WithContext(ctx)
		bufferChan = make(chan *domain.Block, size)
	)

	// 设置并发量
	gg.SetLimit(1000) // TODO: z

	// 并发获取
	for i := uint64(0); i < size; i++ {
		targetBlock := startAt + i
		gg.Go(func() error {
			block, err := srv.fetcher.GetBlockByNumber(gCtx, targetBlock)
			if err != nil {
				return err
			}

			bufferChan <- block
			return nil
		})
	}

	go func() {
		_ = gg.Wait()
		close(bufferChan)
	}()

	var blocks = make([]*domain.Block, 0, size)
	for block := range bufferChan {
		blocks = append(blocks, block)
	}

	if err := gg.Wait(); err != nil {
		return nil, err
	}

	// 对结果进行排序
	sort.Slice(blocks, func(i, j int) bool { return blocks[i].Number < blocks[j].Number })

	return blocks, nil
}

func (srv *IndexDomainService) loadBlockLoop() error {
	helper := log.NewHelper(srv.log)
	helper.Info("start block load loop")
	defer helper.Info("stop block load loop")

	var lastLoadNumber = uint64(0)
	if srv.status.LastSyncBlock != nil {
		lastLoadNumber = srv.status.LastSyncBlock.Number
	}

	for {
		select {
		case <-srv.ctx.Done():
			return nil
		default:
		}

		blocks, err := srv.blockRepo.GetPendingBlocksWithTransactionsByNumber(srv.ctx, lastLoadNumber, 10)
		if err != nil {
			return err
		}

		if len(blocks) == 0 {
			helper.Info("blocks is empty, wait 10 second")
			ticker := time.NewTicker(time.Second * 10)
			select {
			case <-srv.ctx.Done():
				return nil
			case <-ticker.C:
				ticker.Stop()
				ticker = nil
				continue
			}
		}

		for _, block := range blocks {
			lastLoadNumber = block.Number
			select {
			case <-srv.ctx.Done():
				return nil
			case srv.handleQueue <- block:
				//helper.Debugf("send block to handle queue, block number: %d", lastLoadNumber)
			}
		}
	}
}

func (srv *IndexDomainService) handleBlockLoop() error {
	helper := log.NewHelper(srv.log)
	helper.Info("start block handle loop")
	defer helper.Info("stop block handle loop")

	for {
		select {
		case <-srv.ctx.Done():
			return nil

		case block := <-srv.handleQueue:

			// TODO: z debug 功能
			if srv.handleEndBlock != 0 && block.Number > srv.handleEndBlock {
				helper.Infof("block handle done. current_block: %d, end_block: %d", block.Number, srv.handleEndBlock)
				return nil
			}

			if err := srv.handler.HandleBlock(srv.ctx, block); err != nil {
				helper.Errorf("handle block error: %s", err)
				return err
			}

			srv.status.LastSyncBlock = block.Header() // 更新最后同步区块数据
		}
	}
}
