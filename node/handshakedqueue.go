package node

import (
	"github.com/elastos/Elastos.ELA/log"
	"net"
	"sync"
	"time"

	"github.com/elastos/Elastos.ELA/protocol"
)

/*
Handshake queue is a connection queue to handle all connections in
handshake process. When a tcp connection has been started or accepted,
it must finish the handshake progress in less than two seconds, or it
will be disconnected for handshake timeout. And the handshake queue has
a capacity that limited by the DefaultMaxPeers according to P2P protocol.
*/
type handshakeQueue struct {
	sync.Mutex
	capChan chan protocol.Noder
	conns   map[protocol.Noder]net.Conn
}

func (q *handshakeQueue) init() {
	q.capChan = make(chan protocol.Noder, protocol.DefaultMaxPeers)
	q.conns = make(map[protocol.Noder]net.Conn, protocol.DefaultMaxPeers)
}

func (q *handshakeQueue) AddToHandshakeQueue(addr string, node protocol.Noder) {
	log.Info("add to handshake queue 1")
	q.capChan <- node
	log.Info("add to handshake queue 2")
	q.Lock()
	q.conns[node] = node.GetConn()
	q.Unlock()
	log.Info("add to handshake queue 3")
	// Close handshake timeout connections
	go q.handleTimeout(addr, node)
}

func (q *handshakeQueue) RemoveFromHandshakeQueue(node protocol.Noder) {
	q.Lock()
	if _, ok := q.conns[node]; ok {
		delete(q.conns, node)
		<-q.capChan
	}
	q.Unlock()
}

func (q *handshakeQueue) handleTimeout(addr string, node protocol.Noder) {
	time.Sleep(time.Second * protocol.HandshakeTimeout)
	q.Lock()
	log.Info("handle timeout 1")
	if conn, ok := q.conns[node]; ok {
		conn.Close()
		log.Info("handle timeout 2")
		delete(q.conns, node)
		log.Info("handshake queue: remove from connecting list, address:", addr)
		LocalNode.RemoveFromConnectingList(addr)
		<-q.capChan
	}
	q.Unlock()
}
