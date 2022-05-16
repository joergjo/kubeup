# kubeup

`kuebup` is sample webhook written in [Go](https://go.dev) to process Kubernetes Service (AKS) events received from Azure Event Grid using [CloudEvents](https://cloudevents.io). 

Events received by `kubeup` are handled by a `Notifier`. The sample provides two `Notifier` implementations:
- `LogNotifier` writes received events to Go's `log` package's default logger.
- `SendGridNotifier` forwards received events via E-mail using [Twilio SendGrid](https://sendgrid.com). 

## Setup

`kubeup` requires a Azure Event Grid system topic subscription with a webhook endpoint.  

```
cluster_resource_id=$(az aks show -g <resource-group> -n <cluster-name> --query id --output tsv)
az eventgrid event-subscription create --name <subscription-name> \
    --source-resource-id $cluster_resource_id \
    --event-delivery-schema cloudeventschemav1_0 \
    --endpoint <webhook-uri>
```

See [Quickstart: Subscribe to Azure Kubernetes Service (AKS) events with Azure Event Grid (Preview)](https://docs.microsoft.com/en-us/azure/aks/quickstart-event-grid) and [Webhook Event delivery](https://docs.microsoft.com/en-us/azure/event-grid/webhook-event-delivery) if you are not familar with the underlying concepts.

Since Azure Event Grid delivers events only to public endpoints, you must either run `kubeup` on an Azure service that allows you to expose a public endpoint (App Service, Container App, AKS, VMs, etc.), or you can use a reverse proxy service like [ngrok](https://ngrok.com) to route events to a local endpoint.

`kubeup` does _not_ implement any authorization. For a production grade implemetation, you should [secure your webhook endpoint with Azure AD](https://docs.microsoft.com/en-us/azure/event-grid/secure-webhook-delivery).

## Building
Building `kubeup` requires [Go 1.18 or later](https://go.dev/dl/).

```
$ cd kubeup
$ go build -o ./kubeup cmd/main.go 
$ ./kubeup --help
```
Alternatively, you can use the included `Dockerfile` to build a Docker container image. A ready to use container image is available [here](https://hub.docker.com/repository/docker/joergjo/kubeup)

## Configuration
By default, `kubeup` uses the `LogNotifier`. To use `SendGridNotifier`, set the following environment variables before starting `kubeup`:

- `KU_SENDGRID_APIKEY`: Your SendGrid API key
- `KU_SENDGRID_FROM`: The E-mail's from address    
- `KU_SENDGRID_FROM`: The E-mail's to address    
- `KU_SENDGRID_SUBJECT`: The E-mail's subject

All four environment variables must be set to enable the `SendGridNotifier`.

## Testing
To test `kubeup`, you can use the included [sample request](testdata/sample.json). Use any HTTP client like curl, httpie, wget, or Postman to send (HTTP POST) the sample request to the webhook endpoint.
