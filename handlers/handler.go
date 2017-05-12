package handlers

import (
	"github.com/pinshare/config"
	"google.golang.org/grpc"
)

type serviceInterface interface {
	Name() string
	Register(*grpc.Server, *config.Config) error
}

var _handlers = []serviceInterface{}

func addService(service serviceInterface) {
	_handlers = append(_handlers, service)
}

func GetHandlers() []serviceInterface {
	return _handlers
}
