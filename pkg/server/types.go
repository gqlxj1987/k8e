package server

import (
	"context"

	"github.com/xiaods/k8e/pkg/daemons/config"
)

type Config struct {
	DisableAgent     bool
	DisableServiceLB bool
	ControlConfig    config.Control
	Rootless         bool
	SupervisorPort   int
	StartupHooks     []func(context.Context, <-chan struct{}, string) error
}
