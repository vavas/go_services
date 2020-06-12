package gnats

import (
	"sync"

	"github.com/nats-io/nats.go"
)

// lockedConnection is for storing connections with mutex locking.
type lockedConnection struct {
	sync.RWMutex
	bare    *nats.Conn
	encoded *nats.EncodedConn
}

func (lc *lockedConnection) getBare() *nats.Conn {
	lc.RLock()
	defer lc.RUnlock()

	return lc.bare
}

func (lc *lockedConnection) setBare(bare *nats.Conn) {
	lc.Lock()
	defer lc.Unlock()

	lc.bare = bare
}

func (lc *lockedConnection) getEncoded() *nats.EncodedConn {
	lc.RLock()
	defer lc.RUnlock()

	return lc.encoded
}

func (lc *lockedConnection) setEncoded(encoded *nats.EncodedConn) {
	lc.Lock()
	defer lc.Unlock()

	lc.encoded = encoded
}
