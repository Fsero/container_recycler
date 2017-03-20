package handlers

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io/ioutil"
	"os"
	"time"
)

type ContainerInfo struct {
	Name, ID string
}

func ListRunningContainers() []ContainerInfo {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Panic(err)
	}

	container_list := make([]ContainerInfo, 0)
	for _, container := range containers {
		container_list = append(container_list, ContainerInfo{container.Image, container.ID[:10]})
	}
	printContainerList(container_list)
	return container_list
}

func printContainerList(container_list []ContainerInfo) {
	for _, container := range container_list {
		fmt.Printf("%s, %s", container.Name, container.ID)
	}
}

func ScheduleContainerStop(ContainerName string, container_list []ContainerInfo, wait_for_stop *time.Duration) {
	var string alreadyBeingDeletedFlag = "/var/tmp/container_recycler_scheduled_stop"
	fi, err := os.Stat(alreadyBeingDeletedFlag)
	if err == nil {

		log.Infof("container %s already scheduled for stopping", ContainerName)
		return
	}
	log.Infof("scheduled container %s for stopping", ContainerName)
	// creating the flag
	err := ioutil.WriteFile(alreadyBeingDeletedFlag, data, 0644)
	if err != nil {
		log.Panic(err)
	}
	doneChan := make(chan bool)
	time.AfterFunc(*wait_for_stop, func() {
		var data []byte
		data = make([]byte, 1)
		StopContainer(ContainerName, container_list)
		doneChan <- true
	})
	// wait for timer to end
	<-doneChan
}

func StopContainer(ContainerName string, container_list []ContainerInfo) {
	for _, container := range container_list {
		if container.Name == ContainerName {

			cli, err := client.NewEnvClient()
			if err != nil {
				log.Panic(err)
			}
			// We can consider than 10 seconds of timeout is more than enough.
			timeout := (time.Duration(10) * time.Second)
			err = cli.ContainerStop(context.Background(), container.ID, &timeout)
			if err != nil {
				log.Panic(err)
			}
			log.Infof("container %s stopped", ContainerName)
			break
		}
	}

}
