package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/allegro/bigcache"
	"github.com/kevin88886/eth_indexer/internal/domain/balance"
	rctx "github.com/kevin88886/eth_indexer/internal/infrastructure/repository/context"
)

const (
	BalanceCacheKeyPrefix = "balance####"
)

type balanceMemoryRepo struct {
	db    balance.BalanceRepository
	cache *bigcache.BigCache
	mutex sync.Mutex
}

func (repo *balanceMemoryRepo) Save(ctx context.Context, entities ...*balance.Balance) error {
	updateKind := rctx.UpdateKindFromContext(ctx)
	switch updateKind {
	case rctx.UpdateCache:
		return repo.updateCache(entities...)

	case rctx.UpdateDB:
		return repo.db.Save(ctx, entities...)
	default:
		return nil
	}
}

func (repo *balanceMemoryRepo) Load(ctx context.Context, key balance.BalanceKey) (*balance.Balance, error) {

	// 从缓存获取
	entity, err := repo.getCache(key.String())
	if err == nil {
		return entity, nil
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	// 再次从缓存获取
	entity, err = repo.getCache(key.String())
	if err == nil {
		return entity, nil
	}

	// 从数据库获取
	entity, err = repo.db.Load(ctx, key)
	if err != nil {
		return nil, err // database error
	}

	// TODO: z 可以标记不存在的实体, 避免重复请求数据库
	// 设置缓存
	if entity != nil {
		repo.setCache(entity)
	}

	return entity, nil
}

func (repo *balanceMemoryRepo) updateCache(entities ...*balance.Balance) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for _, entity := range entities {
		// id == 0. 说明是新建的. 忽略
		if entity.ID == 0 {
			continue
		}

		repo.setCache(entity)
	}

	return nil
}

func (repo *balanceMemoryRepo) setCache(entity *balance.Balance) {
	bytes, err := entity.Marshal()
	if err != nil {
		return
	}

	key := fmt.Sprintf("%s_%s", BalanceCacheKeyPrefix, entity.Key())
	_ = repo.cache.Set(key, bytes)
}

func (repo *balanceMemoryRepo) getCache(key string) (*balance.Balance, error) {
	bytes, err := repo.cache.Get(fmt.Sprintf("%s_%s", BalanceCacheKeyPrefix, key))
	if err != nil {
		return nil, err
	}

	entity := new(balance.Balance)
	return entity, entity.Unmarshal(bytes)
}

func NewBalanceMemoryRepository(db balance.BalanceRepository, cache *bigcache.BigCache) balance.BalanceRepository {
	return &balanceMemoryRepo{
		db:    db,
		cache: cache,
		mutex: sync.Mutex{},
	}
}
