package xnet

import (
	"context"
)

type Server interface {
	Serve() error
	Shutdown(ctx context.Context) error
	Close() error
}
