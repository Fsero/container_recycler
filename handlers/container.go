package handlers

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"
)

type ContainerInfo struct {
	Name, ID string
}

func ListRunningContainers() []ContainerInfo {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Fatal(err)
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

func deleteFlagFile(alreadyBeingDeletedFlag string, container ContainerInfo) {

	err := os.Remove(alreadyBeingDeletedFlag)
	if err != nil {
		log.Fatalf("unable to delete flag file %s for container %s ID=%s", alreadyBeingDeletedFlag, container.Name, container.ID)
	}
}
func checkIfExistsFlag(alreadyBeingDeletedFlag string, container ContainerInfo) (bool, error) {

	found := false
	fi, err := os.Stat(alreadyBeingDeletedFlag)
	if err == nil {
		found = true
		modtime := fi.ModTime()
		duration := time.Since(modtime)
		if duration.Minutes() > 20 {
			log.Infof("looks like file was not deleted in last execution, cleaning up...")
			found = false
			deleteFlagFile(alreadyBeingDeletedFlag, container)
		}
		log.Debugf("container %s already scheduled for being deleted ", container.Name)
		return found, err
	}

	return found, err

}

func ScheduleContainerStop(ctx context.Context, container ContainerInfo) {

	tmp_prefix_path := ctx.Value("tmp_flags_file_path").(string)
	alreadyBeingDeletedFlag := tmp_prefix_path + container.ID
	flag, err := checkIfExistsFlag(alreadyBeingDeletedFlag, container)
	if flag {
		log.Fatal(err)
	}
	log.Infof("scheduled container %s for stopping", container.Name)
	timeout_duration, err := time.ParseDuration(ctx.Value("exposure_time").(string))
	if err != nil {
		log.Fatalf("incorrect format for exposure_timeout")
	}
	log.Debug("ScheduleContainerStop: outside the lambda function waiting for DONE signal")

	//wait for the exposure_time
	timer := time.NewTimer(timeout_duration)
	runtime.Gosched()
	<-timer.C
	var data []byte
	data = make([]byte, 1)
	// creating the flag
	err = ioutil.WriteFile(alreadyBeingDeletedFlag, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
	StopContainer(ctx, container)
	deleteFlagFile(alreadyBeingDeletedFlag, container)
	log.Debug("ScheduleContainerStop: Lambda function DONE")

}

func GetContainerByName(ContainerName string, container_list []ContainerInfo) (ContainerInfo, bool) {
	for _, container := range container_list {
		if container.Name == ContainerName {
			return container, true
		}
	}
	return ContainerInfo{}, false
}

func GetContainerByID(ContainerID string, container_list []ContainerInfo) (ContainerInfo, bool) {
	if len(ContainerID) <= 0 {
		log.Debug("GetContainerByID: NIL container id provided")
		return ContainerInfo{}, false
	}

	if len(container_list) <= 0 {
		log.Debug("GetContainerByID: NIL container_list provided")
		return ContainerInfo{}, false
	}
	for _, container := range container_list {
		if len(container.ID) < len(ContainerID) {
			log.Debugf("incomparable ID, provided ID is larger than existing one %s %d %s %d", container.ID, len(container.ID), ContainerID, len(ContainerID))
			ContainerID = ContainerID[:len(container.ID)]
		}
		//convert to string
		IDstr := fmt.Sprint(container.ID)
		if strings.HasPrefix(IDstr, ContainerID) {
			return container, true
		}
	}
	return ContainerInfo{}, false
}

func StopContainer(ctx context.Context, container ContainerInfo) {
	log.Infof("Stopping container %s NOW!", container.Name)
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}
	timeout, err := time.ParseDuration(ctx.Value("container_api_timeout").(string))
	if err != nil {
		log.Fatalf("incorrect format for api timeout")
	}
	err = cli.ContainerStop(context.Background(), container.ID, &timeout)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("container %s has been stopped", container.Name)
}
