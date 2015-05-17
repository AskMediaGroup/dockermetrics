package main

import (
	"github.com/fsouza/go-dockerclient"
	"log"
	"sync"
	"time"
)

type Client struct {
	client     *docker.Client
	containers map[string]*docker.Container
	images     map[string]string
	events     chan *docker.APIEvents
	quit       chan struct{}
	formatter  Formatter
	sync.Mutex
}

func NewClient(formatter Formatter) *Client {
	return &Client{
		containers: make(map[string]*docker.Container),
		images:     make(map[string]string),
		events:     make(chan *docker.APIEvents),
		quit:       make(chan struct{}),
		formatter:  formatter,
	}
}

// Connect to docker endpoint
func (c *Client) Connect() {
	client, err := docker.NewClient(flags.DockerEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	c.client = client
	err = client.AddEventListener(c.events)
}

// Start sync ticker
func (c *Client) SyncTick() {

	syncTicker := time.NewTicker(time.Duration(flags.SyncInterval) * time.Second)
	go func() {
		for {
			select {
			case <-syncTicker.C:
				c.Sync()
			case <-c.quit:
				syncTicker.Stop()
				return
			}
		}
	}()
}

// Sync existing containers
func (c *Client) Sync() {
	log.Printf("Syncing existing containers")
	containers, err := c.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, container := range containers {
		c.Add(container.ID)
	}
}

// Watch docker event stream for container start/stop events
func (c *Client) Watch() {
	for msg := range c.events {
		switch msg.Status {
		case "start":
			go c.Add(msg.ID)
		case "die", "stop", "kill":
			go c.Remove(msg.ID)
		}
	}

	close(c.quit)
}

// Start watching a container if we're not already doing so
func (c *Client) Add(id string) {
	c.Lock()
	defer c.Unlock()
	if t := c.containers[id]; t == nil {
		container, err := c.client.InspectContainer(id)
		if err != nil {
			log.Printf("Unable to lookup container\n")
			return
		}
		log.Printf("Adding container %s (%s)", container.Name, id)
		c.containers[id] = container
		go c.Stream(container)
	}

}

// Remove watched container
func (c *Client) Remove(id string) {
	c.Lock()
	defer c.Unlock()
	log.Printf("Removing container %s", id)
	delete(c.containers, id)
}

// Attach to docker stats stream for a container
func (c *Client) Stream(container *docker.Container) {
	id := container.ID
	defer c.Remove(id)
	ch := make(chan *docker.Stats)
	defer close(ch)
	s, r := (chan<- *docker.Stats)(ch), (<-chan *docker.Stats)(ch)
	opts := docker.StatsOptions{ID: id, Stats: s}
	go func() {
		c.client.Stats(opts)
	}()
	for stats := range r {
		c.formatter.Fmt(stats, container, c.ImgName(container.Image))
	}
}

// Lookup image name by ID and cache results
func (c *Client) ImgName(id string) string {
	name := c.images[id]
	if name == "" {
		c.Lock()
		defer c.Unlock()
		images, err := c.client.ListImages(docker.ListImagesOptions{})
		if err != nil {
			log.Printf("Unable to read image name")
			name = id
		} else {
			for _, image := range images {
				if image.ID == id {
					name = image.RepoTags[0]
					if name == "<none>:<none>" {
						name = id[:12]
					}
					break
				}
			}
		}
		c.images[id] = name
	}
	return name
}
