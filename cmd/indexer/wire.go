//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/kevin88886/eth_indexer/internal/conf"
	"github.com/kevin88886/eth_indexer/internal/domain/service"
	"github.com/kevin88886/eth_indexer/internal/facade"
	"github.com/kevin88886/eth_indexer/internal/infrastructure/repository"
)

// wireApp init kratos application.
func wireApp(string, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		conf.ProviderSet,
		repository.ProviderSet,
		service.ProviderSet,
		facade.ProviderSet,
		newApp,
	))
}
