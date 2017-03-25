package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"io"
	"os"
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

func ParseFalcoNotifications(r io.Reader, ctx context.Context) {
	var f FalcoNotification
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := json.Unmarshal(scanner.Bytes(), &f); err != nil {
			log.Error(err)
		} else {
			if f.Priority == "Alert" {
				log.Info("Alert received, will try to stop container")
				container_list := ListRunningContainers()
				timeout_duration, err := time.ParseDuration(ctx.Value("exposure_time").(string))
				if err != nil {
					log.Panic("incorrect timeout set, please specify a right one")
				}
				for _, name := range (ctx.Value("container_images_list_to_stop")).([]string) {
					ScheduleContainerStop(name, container_list, &timeout_duration, ctx)
				}
			}
			log.Info(f)
			log.Info(f.Time.String())
		}
	}
}
