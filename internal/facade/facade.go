package facade

import (
	"time"

	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	pb "github.com/kevin88886/eth_indexer/api/indexer"
	"github.com/kevin88886/eth_indexer/internal/conf"
	"github.com/kevin88886/eth_indexer/internal/facade/handler"
	"github.com/kevin88886/eth_indexer/pkg/middleware"
	"github.com/kevin88886/eth_indexer/pkg/middleware/logging"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/encoding/protojson"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	handler.NewIndexHandler,
	NewGRPCServer,
	NewHTTPServer,
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(conf *conf.Config, h *handler.IndexHandler, logger log.Logger) *grpc.Server {
	c := conf.Bootstrap.Server

	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			validate.Validator(),
		),
		grpc.Options(
			ggrpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             time.Second * 30,
				PermitWithoutStream: true,
			}),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}

	srv := grpc.NewServer(opts...)
	pb.RegisterIndexerServer(srv, h)
	return srv
}

func init() {
	// 重置 kratos encoding 的 marshal 配置
	json.MarshalOptions = protojson.MarshalOptions{
		Multiline:       false,
		Indent:          "",
		AllowPartial:    false,
		UseProtoNames:   false,
		UseEnumNumbers:  true,
		EmitUnpopulated: true,
		Resolver:        nil,
	}
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(config *conf.Config, h *handler.IndexHandler, logger log.Logger) *http.Server {
	c := config.Server

	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			middleware.Cors(),
			logging.Server(logger),
			validate.Validator(),
		),
		http.Timeout(time.Second * 30),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)
	pb.RegisterIndexerHTTPServer(srv, h)
	return srv
}
