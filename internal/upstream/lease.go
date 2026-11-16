package upstream

import "sync"

type Lease struct {
	Endpoint *Endpoint

	once    sync.Once
	release func()
}

func NewLease(endpoint *Endpoint, release func()) *Lease {
	return &Lease{
		Endpoint: endpoint,
		release:  release,
	}
}

func (l *Lease) Release() {
	if l == nil {
		return
	}

	l.once.Do(func() {
		if l.release != nil {
			l.release()
		}
	})
}
