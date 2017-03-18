package main

import (
	"bitbucket.org/fseros/container_recycler/handlers"
	"os"
	"strings"
)

//{"output":"17:20:45.212076717: Alert Shell spawned in a container other than entrypoint)",
// "priority":"Alert","rule":"Run shell in container","time":"2017-02-26T17:20:45.212076717Z"}

func main() {
	//scanner := bufio.NewScanner(os.Stdin)
	//for scanner.Scan() {
	//	fmt.Println(scanner.Text())
	//}
	//b := []byte(`{"output":"17:20:45.212076717: Alert Shell spawned in a container other than entrypoint)","priority":"Alert","rule":"Run shell in container","time":"2017-02-26T17:20:45.212076717Z"}`)
	handlers.SetupLogging()
	for _, arg := range os.Args[1:] {
		r := strings.NewReader(arg)
		handlers.CheckNotifications(r)
	}
}
