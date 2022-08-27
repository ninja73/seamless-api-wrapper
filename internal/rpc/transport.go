package rpc

import (
	"context"
	"io"
)

type Transport interface {
	Run(ctx context.Context, resolver Resolver) error
}

type Resolver interface {
	Resolve(ctx context.Context, writer io.Writer, reader io.Reader)
}
