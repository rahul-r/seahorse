package containers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ContainerInfo struct {
	Name        string
	TemplateDir string
	ID          string `json:"Id"`
	Image       string
	ImageID     string
	Command     string
	Created     int64
	SizeRw      int64 `json:",omitempty"`
	SizeRootFs  int64 `json:",omitempty"`
	Labels      map[string]string
	State       string
	Status      string
	HostConfig  struct {
		NetworkMode string            `json:",omitempty"`
		Annotations map[string]string `json:",omitempty"`
	}
	// Names      []string
	// Ports      []Port
	// NetworkSettings *SummaryNetworkSettings
	// Mounts          []MountPoint
}

type ComposeMap map[string]ContainerInfo

type Containers struct {
	client       *client.Client
	containerMap ComposeMap
}

func NewLocalClient() Containers {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	return Containers{client, make(ComposeMap)}
}

// host format: ssh://[user]@[host]:[port]
func NewRemoteClient(host string) Containers {
	helper, err := connhelper.GetConnectionHelper(host)
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{
		// No tls
		// No proxy
		Transport: &http.Transport{
			DialContext: helper.Dialer,
		},
	}

	var clientOpts []client.Opt

	clientOpts = append(clientOpts,
		client.WithHTTPClient(httpClient),
		client.WithHost(helper.Host),
		client.WithDialContext(helper.Dialer),
	)

	version := os.Getenv("DOCKER_API_VERSION")

	if version != "" {
		clientOpts = append(clientOpts, client.WithVersion(version))
	} else {
		clientOpts = append(clientOpts, client.WithAPIVersionNegotiation())
	}

	client, err := client.NewClientWithOpts(clientOpts...)
	if err != nil {
		panic(err)
	}

	return Containers{client, make(ComposeMap)}
}

func (self *Containers) GetContainerMap() *ComposeMap {
	return &self.containerMap
}

func (self *Containers) GetContainerStatus(containerName string) string {
	container, err := self.client.ContainerInspect(context.Background(), containerName)
	if err != nil {
		panic(err)
	}
	return container.State.Status
}

func (self *Containers) CreateContainerMap(input ComposeMap) error {
	containers, err := self.client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		log.Fatal(err)
		return err
	}

	for _, ctr := range containers {
		name := strings.TrimLeft(ctr.Names[0], "/")

		if entry, ok := input[name]; ok {
			// key exists in the map
			entry.ID = ctr.ID
			entry.Image = ctr.Image
			entry.State = ctr.State
			input[name] = entry
		} else {
			// key does not exist in the map
			input[name] = ContainerInfo{
				ID:    ctr.ID,
				Image: ctr.Image,
				State: ctr.State,
				Name:  name,
			}
		}
	}

	self.containerMap = input

	return nil
}

func (self *Containers) Start(name string) error {
	ctx := context.Background()
	containers, err := self.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return err
	}

	containerID := ""
	for _, ctr := range containers {
		if strings.TrimLeft(ctr.Names[0], "/") == name {
			containerID = ctr.ID
		}
	}

	if containerID == "" {
		return errors.New(fmt.Sprintf("Cannot find container `%s`", name))
	}

	log.Printf("Starting %s", name)

	self.client.ContainerStart(ctx, containerID, container.StartOptions{})

	return nil
}

func (self *Containers) Stop(name string) error {
	ctx := context.Background()
	containers, err := self.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return err
	}

	containerID := ""
	for _, ctr := range containers {
		if strings.TrimLeft(ctr.Names[0], "/") == name {
			containerID = ctr.ID
		}
	}

	if containerID == "" {
		return errors.New(fmt.Sprintf("Cannot find container `%s`", name))
	}

	waitTimeoutSec := 20

	log.Printf("Stopping %s", name)

	self.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &waitTimeoutSec})

	return nil
}
