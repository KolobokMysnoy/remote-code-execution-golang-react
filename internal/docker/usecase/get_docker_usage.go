package usecase

import (
	"check_system/internal/domain"
	"context"
	"time"
)

type DockerSystemUsecase struct {
	pool *DockerConnectionPool

}


// TODO Make config
func NewDockerSystem(minConnections, maxConnections int, 
	timeout time.Duration, 
	languages map[string]string, 
	ctx context.Context,
	) (*DockerSystemUsecase, error) {
	dockerSystem := &DockerSystemUsecase{}

	pool, err := NewDockerConnectionPool(minConnections, maxConnections, timeout, languages, ctx) 
	if err != nil{
		return nil, err
	}
	dockerSystem.pool = pool
	return dockerSystem, nil
}

func (d *DockerSystemUsecase) GetContainer(language string, ctx context.Context) (domain.Runner, error) {
	conn, err := d.pool.GetConnection(language)
	if err!= nil {
		return nil, err
	}

	releaseFunc := func() {
		d.pool.ReleaseConnection(language, conn)
	}

	runner, err := TransformConnToRunner(conn, language, releaseFunc)
	return runner, nil
}

func (d *DockerSystemUsecase) SetMaxContainers(maxCont int, ctx context.Context) (error) {
	d.pool.maxPoolSize = maxCont
	return nil
}

func (d *DockerSystemUsecase) SetMinContainers(minCont int, ctx context.Context) (error) {
	d.pool.minPoolSize = minCont
	return nil
}
