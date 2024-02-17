package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	consts "check_system/config"

	"github.com/docker/docker/client"
)

type DockerConnectionPool struct {
	mu sync.Mutex
	connections map[string]*connectionInfo
	maxPoolSize int
	minPoolSize int
	Timeout time.Duration
}

func NewDockerConnectionPool(minConnections, maxConnections int, timeout time.Duration, languages map[string]string, ctx context.Context) (*DockerConnectionPool, error) {
	if minConnections > maxConnections {
		return nil, fmt.Errorf("max connections is less than min connection")
	}


	dockerPool := &DockerConnectionPool{
		maxPoolSize: maxConnections,
		minPoolSize: minConnections,
		connections: make(map[string]*connectionInfo, consts.LanguagesLen),
		Timeout: timeout,
	}

	go dockerPool.PeriodRelease(ctx)

	return dockerPool, nil
} 

func (p *DockerConnectionPool) createDockerClient(language string) (*client.Client, error) {
	// TODO pass language that need to create docker container
	cli, err := client.NewClientWithOpts()
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (p *DockerConnectionPool) createLanguageInfo(language string, maxPoolSize int) error {
	p.connections[language] = &connectionInfo{
			ch: make(chan *client.Client, p.maxPoolSize),
		}
	return nil
}

func (p *DockerConnectionPool) GetConnection(language string) (*client.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.connections[language]; !ok {
		err := p.createLanguageInfo(language, p.maxPoolSize)
		if err != nil {
			return nil, err
		}
	}

	info := p.connections[language]
	timer := time.NewTimer(consts.WaitForContainer * time.Second)
	defer timer.Stop()

	select {
	case conn:= <-info.GetConnection():
		return conn, nil
	default:
		if len(info.ch) < p.maxPoolSize {
			conn, err := info.AddConnection(func() (*client.Client, error) {
				return p.createDockerClient(language)
			})	
			if err != nil {
				return nil, err
			}
			
			return conn, nil 
		}
		
		// TODO fix this
		select {
		case conn:= <-info.GetConnection():
			timer.Stop()
			return conn, nil
		case <-timer.C:
			return nil, consts.ErrorChannelFull
		}
	}
}

func (p *DockerConnectionPool) ReleaseConnection(language string, conn *client.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()

	info, ok := p.connections[language]
	if !ok {
		return 
	}

	info.ReturnConnection(conn)
}

// Run in gorutine to periodical check. If run not in gorutine will crash app!
func (p* DockerConnectionPool) PeriodRelease(ctx context.Context) {
	timer := time.NewTimer(consts.PeriodForContainerCheck * time.Second)
	defer timer.Stop()

	Loop:
	for {
		select {
		case <-timer.C:
			p.mu.Lock()
			defer p.mu.Unlock()

			for _, con := range p.connections {
				con.ConnectionRelease(p.minPoolSize)
			}
			
			timer.Reset(0)
		case <-ctx.Done():
			break Loop
		}
		
	}
}