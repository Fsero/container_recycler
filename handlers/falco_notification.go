package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"io"
	"os"
	"regexp"
	"sync"
	"time"
)

type FalcoNotification struct {
	RawOutput         string    `json:"output"`
	Priority          string    `json:"priority"`
	RuleNameTriggered string    `json:"rule"`
	Time              time.Time `json:"time"`
}

func SetupLogging() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
}

//{"output":"14:54:07.709160152: Alert Shell spawned in a container other than entrypoint (user=root ssh (id=52d928d8b2a3) shell=sh parent=watch cmdline=sh -c id)","priority":"Alert","rule":"Run shell in container","time":"2017-03-31T14:54:07.709160152Z"}

func handle(ctx context.Context, f FalcoNotification, wg sync.WaitGroup) {

	log.Info(f)
	log.Info(f.Time.String())

	if f.Priority == "Alert" {

		var myExp = namedRegexp{regexp.MustCompile(`.*\(user=(?P<user>[[:alpha:]]+)\s+(?P<image_name>[[:alpha:]]+)\s+\(id=(?P<image_id>[[:alnum:]]{6,})\).*\)`)}
		data := myExp.FindStringSubmatchMap(f.RawOutput)
		log.Info(data)

		log.Debug("Alert received, will try to stop container")
		container_list := ListRunningContainers()

		ctx = context.WithValue(ctx, "container_list", container_list)
		container, found := GetContainerByID(data["image_id"], container_list)
		if found {
			log.Debug("FalcoNotification.handle: stopping container")
			ScheduleContainerStop(ctx, container)
		} else {
			log.Warnf("Alert received relative for container ID %s (name=%s) not found in running containers", data["image_id"], data["image_name"])
		}
	}
}

func ParseFalcoNotifications(r io.Reader, ctx context.Context) {
	var f FalcoNotification
	scanner := bufio.NewScanner(r)
	var wg sync.WaitGroup
	for scanner.Scan() {
		if err := json.Unmarshal(scanner.Bytes(), &f); err != nil {
			log.Error(err)

			log.Debug("ParseFalcoNotifications: Bad FalcoNotification format")
			continue
		}
		wg.Add(1)
		log.Debug("ParseFalcoNotifications: received a falco notification")

		go func() {
			defer wg.Done()
			handle(ctx, f, wg)
		}()
	}
	wg.Wait()
}
