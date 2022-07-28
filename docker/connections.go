package docker

import (
	"io"
	"log"
	"sync"

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
		pool sync.Pool
	}

	pooledConn struct {
		Connection
		pool *sync.Pool
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
	var pool sync.Pool
	pool.New = func() any {
		conn, err := backend.CreateConn()
		if err != nil {
			panic(err)
		}
		return &pooledConn{conn, &pool}
	}
	return &ConnectionPool{pool: pool}
}

func (p *ConnectionPool) CreateConn() (conn Connection, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if panicErr, ok := recovered.(error); ok {
				err = panicErr
			} else {
				panic(recovered)
			}
		}
	}()
	conn = p.pool.Get().(Connection)
	log.Println("Got connection from pool")
	return
}

func (c *pooledConn) Close() error {
	c.pool.Put(c.Connection)
	log.Println("Returned connection to pool")
	return nil
}
