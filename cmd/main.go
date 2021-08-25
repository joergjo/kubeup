package main

import (
	"context"
	"flag"
	"log"

	"github.com/joergjo/go-samples/kubeup"
)

func main() {
	port := flag.Int("port", 5000, "HTTP listen port")
	path := flag.String("path", "/webhook", "WebHook path")
	flag.Parse()

	err := kubeup.Run(context.Background(), *path, *port)
	if err != nil {
		log.Fatalf("Fatal error while running CloudEvent receiver: %v\n", err)
	}
}
