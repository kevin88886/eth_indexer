package tick

import (
	"github.com/kevin88886/eth_indexer/internal/domain/protocol"
)

type Tick interface {
	GetID() int64                   // tick ID
	GetName() string                // tick 名称
	GetProtocol() protocol.Protocol // tick 所属协议
	LastUpdatedBlock() uint64
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

var (
	_ Tick = (*IERC20Tick)(nil)
	_ Tick = (*IERCPoWTick)(nil)
)
