package usecase

import (
	consts "check_system/config"
	"check_system/internal/domain"
	"context"
	"log"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)



type ConnectionsPool struct {
	conn chan *client.Client 
	mu sync.Mutex
	activeContainer int
	image string
}

// Create connection pool
func NewConnectionsPool(image string, sizeOfPool int) (domain.DockerPool, error) {
	pool := &ConnectionsPool{
		image: image,
		conn: make(chan *client.Client, sizeOfPool),
	}

	return pool, nil
}

func (c *ConnectionsPool) Size() int {
	return len(c.conn)
}

func (c *ConnectionsPool) Active() int {
	return c.activeContainer
}

func (c *ConnectionsPool) create() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return cli, nil
}

// Create and start container in client based on image.
// Return id of container and error if something happen
func (c *ConnectionsPool) createContainer(cli *client.Client, image string) (id string, err error) {
	log.Default().Print("Create container with image: ", image)
	
	ctx := context.Background()
	_, err = cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		log.Default().Print("Cant download image")
		return "", err
	}
	log.Default().Print("Image downloads finished")

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin: true,
		OpenStdin: true,
	}, nil, nil, nil, "")
	if err != nil {
		return "", nil
	}
	
	log.Default().Print("Start container")
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", nil
	}

	return resp.ID, nil
}

// Create connection and container in it
func (c *ConnectionsPool) AddContainer() (error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := c.create()
	if err != nil {
		return err
	}
	log.Default().Print("Add to connection")
	c.conn <- conn

	_, err = c.createContainer(conn, c.image)
	if err != nil {
		log.Default().Print("Error from image container", err)
		return err
	}

	return nil
}

// Release connections and stop containers if don't use it
func (c *ConnectionsPool) ReleaseExtraContainers(desiredCount int) (error) {
	// XXX not run anywhere...
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Default().Print("Release containers")

	if len(c.conn) > desiredCount && len(c.conn) != c.activeContainer {
		for !(len(c.conn) == c.activeContainer || len(c.conn) <= desiredCount) {
			conn:= <-c.conn
			
			containerInfo, err := conn.ContainerList(context.Background(), container.ListOptions{})
			if err != nil {
				return err
			}

			id := containerInfo[len(containerInfo)-1].ID
			conn.ContainerStop(context.Background(), id, container.StopOptions{})

			conn.Close()
		}
	}

	return nil
}

// Return container to pool to use
func (c *ConnectionsPool) ReturnContainer(cli *client.Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.activeContainer -= 1
	c.conn <- cli
}

// Return client with container and get in timeout
// If in timeout can't get container ther return ErrorWaitingContainer
func (c *ConnectionsPool) GetContainer(timeout int) (*client.Client, error) {
	timer:= time.NewTimer(time.Duration(timeout) * time.Second)
	select {
	case <-timer.C:
		return nil, consts.ErrorWaitingContainer
	case con := <-c.conn:
		c.activeContainer += 1
		return con, nil
	}
}

// Set image that will be passed to all containers that will be created in pool
func (c *ConnectionsPool) SetImage(image string) {
	c.image = image
}
