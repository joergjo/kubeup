package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/joergjo/go-samples/kubeup"
)

func main() {
	port := flag.Int("port", 8000, "HTTP listen port")
	path := flag.String("path", "/webhook", "WebHook path")
	flag.Parse()

	var apiKey, from, to, subject string
	var notifier kubeup.Notifier = kubeup.LogNotifier{}
	if getConfigFromEnv(&apiKey, &from, &to, &subject) {
		notifier = kubeup.NewSendGridNotifier(apiKey, from, to, subject)
	}

	err := kubeup.Run(context.Background(), *path, *port, notifier)
	if err != nil {
		log.Fatalf("Fatal error while running CloudEvent receiver: %v", err)
	}
}

func getConfigFromEnv(apiKey, from, to, sub *string) bool {
	var ok bool
	*apiKey, ok = os.LookupEnv("KU_SENDGRID_APIKEY")
	if !ok {
		return false
	}
	*from, ok = os.LookupEnv("KU_SENDGRID_FROM")
	if !ok {
		return false
	}
	*to, ok = os.LookupEnv("KU_SENDGRID_TO")
	if !ok {
		return false
	}
	*sub, ok = os.LookupEnv("KU_SENDGRID_SUBJECT")
	return ok
}
