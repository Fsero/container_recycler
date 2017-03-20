package handlers

import (
	"bufio"
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

func ParseFalcoNotifications(r io.Reader) {
	var f FalcoNotification
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := json.Unmarshal(scanner.Bytes(), &f); err != nil {
			log.Error(err)
		} else {
			if f.Priority == "Alert" {
				log.Info("Alert received")
			}
			log.Info(f)
			log.Info(f.Time.String())
		}
	}
}
