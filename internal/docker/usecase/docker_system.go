package usecase

import (
	consts "check_system/config"
	"check_system/internal/domain"
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)



func NewDockerSystem(languages map[string]string, ctx context.Context) (domain.DockerSystem, error) {
	l := ctx.Value(consts.LoggerCtxName)
	log, ok := l.(*zap.Logger);
    if !ok {
        return nil, fmt.Errorf("logger not exist in context")
    }
	
	system := &DockersSystem{
		languagesImages: languages,
		dockersPool: make(map[string]domain.DockerPool),
		max: 5,
		min: 2,
	}
	
	log.Info("create connection pools")
	// TODO make const
	for language, image := range languages {
		pool, err := NewConnectionsPool(image, 5)
		if err != nil {
			return nil, err
		}

		system.dockersPool[language] = pool
	}

	
	log.Info("run release func")
	go func ()  {
			// TODO const
			ticker := time.NewTicker(60 * time.Second)

			for {
				select {
				case <-ticker.C:
					for _, dp := range system.dockersPool {
						dp.ReleaseExtraContainers(2)
					}
				}
			}
		}()

	return system, nil
}

type DockersSystem struct {
	dockersPool 	map[string]domain.DockerPool
	languagesImages map[string]string
	mu sync.Mutex
	max int
	min int
}

func (d *DockersSystem)SetMax(max int) {
	d.max = max
}

func (d *DockersSystem)SetMin(min int) {
	d.min = min
}

func (d *DockersSystem) GetRunner(language string, ctx context.Context) (domain.Runner, error) {
	// TODO
	l := ctx.Value(consts.LoggerCtxName)
	log, ok := l.(*zap.Logger);
    if !ok {
        return nil, fmt.Errorf("logger not exist in context")
    }

	d.mu.Lock()
	defer d.mu.Unlock()

	pool := d.dockersPool[language]
	if pool.Active() == d.max {
		log.Warn("Containers pool full", zap.String("language", language))
		return nil, fmt.Errorf("containers pool full")
	} 

	if pool.Active() >= pool.Size() {
		log.Info("Create new container", zap.String("language", language))
		pool.AddContainer()
	} 
	log.Info("Container get with timeout")

	// TODO to const
	container, err := d.dockersPool[language].GetContainer(10)
	if err != nil {
		return nil, err
	}

	log.Info("Create new docker runner")

	runner, err := NewDockerRunner(language, container, func ()  {
		log.Info("Release function using")
		d.dockersPool[language].ReturnContainer(container)
	})
	if err != nil {
		return nil, err
	}

	return runner, nil
}

func (d *DockersSystem) ReturnRunner(runner domain.Runner) {
	runner.CloseRunner()
}

