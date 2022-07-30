package docker

import (
	"io"
	"log"
	"sync"
	"time"

	"github.com/docker/docker/client"
)

type (
	Connection interface {
		client.ContainerAPIClient
		client.NetworkAPIClient
		client.VolumeAPIClient
		client.SystemAPIClient
		io.Closer
	}

	ConnectionFactory interface {
		CreateConn() (Connection, error)
	}

	BasicConnectionFactory []client.Opt

	ConnectionPool struct {
		backend ConnectionFactory
		pool    chan *pooledConn
	}

	pooledConn struct {
		Connection
		pool        *ConnectionPool
		idleTimeout *time.Timer
		closed      bool
		mu          sync.Locker
	}
)

var (
	_ ConnectionFactory = (BasicConnectionFactory)(nil)
	_ ConnectionFactory = (*ConnectionPool)(nil)

	_ Connection = (*client.Client)(nil)
	_ Connection = (*pooledConn)(nil)
)

func MakeBasicConnectionFactory(opts ...client.Opt) BasicConnectionFactory {
	return BasicConnectionFactory(opts)
}

func (f BasicConnectionFactory) CreateConn() (Connection, error) {
	client, err := client.NewClientWithOpts(f...)
	if err != nil {
		return nil, err
	}
	log.Println("Opened connection")
	return client, nil
}

func NewConnectionPool(backend ConnectionFactory) *ConnectionPool {
	return &ConnectionPool{backend, make(chan *pooledConn, 20)}
}

func (p *ConnectionPool) CreateConn() (Connection, error) {
	if conn, err := p.getOrCreate(); err == nil {
		conn.acquired()
		return conn, nil
	} else {
		return nil, err
	}
}

func (p *ConnectionPool) getOrCreate() (*pooledConn, error) {
	for {
		select {
		case conn := <-p.pool:
			if !conn.isClosed() {
				return conn, nil
			}
		default:
			log.Println("Creating a new connection")
			if conn, err := p.backend.CreateConn(); err == nil {
				pconn := &pooledConn{
					Connection: conn,
					pool:       p,
					mu:         &sync.Mutex{},
				}
				pconn.idleTimeout = time.AfterFunc(20*time.Hour, pconn.doClose)
				return pconn, nil
			} else {
				return nil, err
			}
		}
	}
}

func (c *pooledConn) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

func (c *pooledConn) acquired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.idleTimeout.Stop() {
		<-c.idleTimeout.C
	}
	log.Println("Connection acquired")
}

func (c *pooledConn) released() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.idleTimeout.Reset(10 * time.Second)
	log.Println("Connection released")
}

func (c *pooledConn) doClose() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		log.Println("Closing connection")
		_ = c.Connection.Close()
		c.closed = true
	}
}

func (c *pooledConn) Close() error {
	select {
	case c.pool.pool <- c:
		c.released()
	default:
		c.doClose()
	}
	return nil
}
