package managed

import (
	"context"
	"net/http"

	"github.com/xiaods/k8e/pkg/clientaccess"
	"github.com/xiaods/k8e/pkg/daemons/config"
)

var (
	defaultDriver string
	drivers       []Driver
)

type Driver interface {
	IsInitialized(ctx context.Context, config *config.Control) (bool, error)
	Register(ctx context.Context, config *config.Control, handler http.Handler) (http.Handler, error)
	Reset(ctx context.Context) error
	Start(ctx context.Context, clientAccessInfo *clientaccess.Info) error
	Test(ctx context.Context) error
	Restore(ctx context.Context) error
	EndpointName() string
}

func RegisterDriver(d Driver) {
	drivers = append(drivers, d)
}

func Registered() []Driver {
	return drivers
}

func Default() string {
	if defaultDriver == "" && len(drivers) == 1 {
		return drivers[0].EndpointName()
	}
	return defaultDriver
}
