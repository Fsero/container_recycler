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
	return container_list
}

func printContainerList(container_list []ContainerInfo) {
	for _, container := range container_list {
		fmt.Printf("%s, %s\n", container.Name, container.ID)
	}
}

func ScheduleContainerStop(ContainerName string, container_list []ContainerInfo, wait_for_stop *time.Duration, ctx context.Context) {
	container, found := GetContainerByName(ContainerName, container_list)
	if !found {
		log.Fatalf("Container %s is not running now, doing nothing", ContainerName)
		return
	}

	var alreadyBeingDeletedFlag string = "/var/tmp/container_recycler_" + container.ID
	fi, err := os.Stat(alreadyBeingDeletedFlag)
	if err == nil {
		modtime := fi.ModTime()
		duration := time.Since(modtime)
		if duration.Minutes() > 20 {
			log.Infof("looks like file was not deleted in last execution, cleaning up...")
			err := os.Remove(alreadyBeingDeletedFlag)
			if err != nil {
				log.Panicf("unable to delete flag file %s", alreadyBeingDeletedFlag)
			}

		}
	}
	log.Infof("scheduled container %s for stopping", ContainerName)
	doneChan := make(chan bool)
	time.AfterFunc(*wait_for_stop, func() {
		var data []byte
		data = make([]byte, 1)
		// creating the flag
		err = ioutil.WriteFile(alreadyBeingDeletedFlag, data, 0644)
		if err != nil {
			log.Panic(err)
		}
		StopContainer(ContainerName, container_list, ctx)
		err = os.Remove(alreadyBeingDeletedFlag)
		if err != nil {
			log.Fatalf("unable to delete flag file %s", alreadyBeingDeletedFlag)
		}
		doneChan <- true
	})
	// wait for timer to end
	<-doneChan
}

func GetContainerByName(ContainerName string, container_list []ContainerInfo) (ContainerInfo, bool) {
	for _, container := range container_list {
		if container.Name == ContainerName {
			return container, true
		}
	}
	return ContainerInfo{}, false
}

func StopContainer(ContainerName string, container_list []ContainerInfo, ctx context.Context) {
	container, found := GetContainerByName(ContainerName, container_list)
	if found {
		cli, err := client.NewEnvClient()
		if err != nil {
			log.Panic(err)
		}
		// We can consider than 10 seconds of timeout is more than enough.
		timeout, err := time.ParseDuration(ctx.Value("container_api_timeout").(string))
		if err != nil {
			log.Fatalf("incorrect format for api timeout")
		}
		err = cli.ContainerStop(context.Background(), container.ID, &timeout)
		if err != nil {
			log.Panic(err)
		}
		log.Infof("container %s stopped", ContainerName)
	} else {
		log.Fatalf("container %s not found running, doing nothing", ContainerName)
		return
	}
}
