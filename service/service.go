package service

import (
	"github.com/pinshare/config"
	"google.golang.org/grpc"
)

type serviceInterface interface {
	Name() string
	Register(*grpc.Server, *config.Config) error
}

var _services = []serviceInterface{}

func addService(service serviceInterface) {
	__services = append(_services, service)
}

func GetServices() []serviceInterface {
	return _services
}
