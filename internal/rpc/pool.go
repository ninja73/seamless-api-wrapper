package rpc

import (
	"sync"
)

type reqPool struct {
	pool sync.Pool
}

func (p *reqPool) get() []byte {
	msg := p.pool.Get()
	if msg == nil {
		return make([]byte, 4096)
	}
	return msg.([]byte)
}

func (p *reqPool) put(data []byte) {
	p.pool.Put(data)
}
