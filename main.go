package main

import (
	"bitbucket.org/fseros/container_recycler/handlers"
	"bufio"
	"context"
	"os"
	"strings"
)

//{"output":"17:20:45.212076717: Alert Shell spawned in a container other than entrypoint)",
// "priority":"Alert","rule":"Run shell in container","time":"2017-02-26T17:20:45.212076717Z"}
// TODO: Replace with config file

func main() {
	handlers.SetupLogging()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = context.WithValue(ctx, "exposure_time", "10s")
	ctx = context.WithValue(ctx, "container_images_list_to_stop", []string{"ssh"})
	ctx = context.WithValue(ctx, "container_api_timeout", "10s")

	// reading arguments

	for _, arg := range os.Args[1:] {
		r := strings.NewReader(arg)

		handlers.ParseFalcoNotifications(r, ctx)
	}

	// read from stdin
	r := bufio.NewReader(os.Stdin)
	handlers.ParseFalcoNotifications(r, ctx)

}
