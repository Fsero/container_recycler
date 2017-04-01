package main

import (
	"bitbucket.org/fseros/container_recycler/handlers"
	"bufio"
	"context"
	"os"
	"runtime"
	"strings"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	handlers.SetupLogging()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: Replace with config file
	ctx = context.WithValue(ctx, "exposure_time", "10m")
	ctx = context.WithValue(ctx, "container_api_timeout", "10s")
	ctx = context.WithValue(ctx, "tmp_flags_file_path", "/var/tmp/container_recycler_")

	// reading arguments

	for _, arg := range os.Args[1:] {
		r := strings.NewReader(arg)
		handlers.ParseFalcoNotifications(r, ctx)
	}

	// read from stdin
	r := bufio.NewReader(os.Stdin)
	handlers.ParseFalcoNotifications(r, ctx)

}
