package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/joergjo/go-samples/kubeup"
)

func main() {
	port := flag.Int("port", 5000, "HTTP listen port")
	path := flag.String("path", "/webhook", "WebHook path")
	flag.Parse()

	notifier := getNotifier()
	err := kubeup.Run(context.Background(), *path, *port, notifier)
	if err != nil {
		log.Fatalf("Fatal error while running CloudEvent receiver: %v", err)
	}
}

func getNotifier() kubeup.Notifier {
	apiKey, ok := os.LookupEnv("KU_SENDGRID_APIKEY")
	if !ok {
		return kubeup.LogNotifier{}
	}
	from, ok := os.LookupEnv("KU_SENDGRID_FROM")
	if !ok {
		return kubeup.LogNotifier{}
	}
	to, ok := os.LookupEnv("KU_SENDGRID_TO")
	if !ok {
		return kubeup.LogNotifier{}
	}
	sub, ok := os.LookupEnv("KU_SENDGRID_SUBJECT")
	if !ok {
		return kubeup.LogNotifier{}
	}
	return kubeup.NewSendGridNotifier(apiKey, from, to, sub)
}
