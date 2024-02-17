package usecase

import (
	"sync"

	"github.com/docker/docker/client"
)

// TODO create containers here

type CreateConnection func() (*client.Client, error)

type connectionInfo struct {
	ch chan *client.Client
	mu sync.Mutex
	currentConnCount int
}

// Will close all connections that inactive and that more than maxConn
func (c *connectionInfo) ConnectionRelease(maxConn int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if len(c.ch) > maxConn && c.currentConnCount < maxConn {
		for c.currentConnCount > maxConn {
			conn:= <-c.ch
			conn.Close()
			c.currentConnCount -= 1
		}
	}
}

// Add connection to connectionInfo and return this connection
func (c *connectionInfo) AddConnection(create CreateConnection) (*client.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err :=create()
	if err != nil {
		return nil, err
	}

	c.ch <- conn
	c.currentConnCount += 1
	
	return <-c.ch, nil
}

func (c *connectionInfo) GetConnection() (chan *client.Client) {
	return c.ch
}

// Return connection to connectionInfo channel
func (c *connectionInfo) ReturnConnection(con *client.Client) {
	c.ch <- con
}