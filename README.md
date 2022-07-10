# kubeup

`kubeup` is a sample WebHook written in [Go](https://go.dev) to process Azure Kubernetes Service (AKS) [CloudEvents](https://cloudevents.io) that notify receivers of new Kubernetes versions being available. Refer to [Quickstart: Subscribe to Azure Kubernetes Service (AKS) events with Azure Event Grid (Preview)](https://docs.microsoft.com/en-us/azure/aks/quickstart-event-grid) and [WebHook Event delivery](https://docs.microsoft.com/en-us/azure/event-grid/webhook-event-delivery) if you want to learn more about the underlying concepts.

Events received by `kubeup` are handled by a `Notifier`. The sample provides two `Notifier` implementations:

- `LogNotifier` writes received events to stderr using [zerolog](github.com/rs/zerolog).
- `SendGridNotifier` forwards received events by E-mail using [Twilio SendGrid](https://sendgrid.com).

`kubeup` does _not_ implement any authorization (yet). For a production grade implemetation, you should [secure your WebHook endpoint with Azure AD](https://docs.microsoft.com/en-us/azure/event-grid/secure-webhook-delivery).

Since Azure Event Grid delivers events only to public endpoints, you must either run `kubeup` on an Azure service that allows you to expose a public endpoint (App Service, Container App, AKS, VMs, etc.), or you can use a reverse proxy service like [ngrok](https://ngrok.com) to route events to a local endpoint. This repo includes Bicep templates to deploy `kubeup` as an [Azure Container App](https://docs.microsoft.com/en-us/azure/container-apps/overview), including [HTTP scaling rules to scale to zero](https://docs.microsoft.com/en-us/azure/container-apps/scale-app).

## Prequisites

You'll need an Azure subscription and a very small set of tools and skills to get started:

1. An Azure subscription. Sign up [for free](https://azure.microsoft.com/free/).
2. Either the [Azure CLI](https://docs.microsoft.com/cli/azure/install-azure-cli) installed locally, or the [Azure Cloud Shell](https://shell.azure.com) available online.
3. If you are using a local installation of the Azure CLI:
   1. You need a bash shell to execute the included deployment script - on Windows 10/11 use the [Window Subsystem for Linux](https://docs.microsoft.com/en-us/windows/wsl/install).
   2. Make sure to have Bicep CLI installed by running `az bicep install`
4. A Twilio SendGrid account. Sign up [for free](https://sendgrid.com/pricing/).

## Deployment

Use the included deployment script to deploy `kubeup` to an Azure Container App that uses the `SendGridNotifier`. The Bicep templates will both deploy a `kubeup` Azure Container App and create a WebHook subscription for an _existing_ AKS cluster&mdash;they will however not create a new cluster.

```bash
$ export KU_RESOURCE_GROUP=my-kubeup-rg
$ cd kubeup
$ ./deploy.sh
```

All resources are created in the same region. You can override the default settings
of the deployment script by exporting the following environment variables. Note that all
environment variables with no default value are required.

| Environment variable     | Purpose                              | Default value           |
| ------------------------ | ------------------------------------ | ----------------------- |
| `KU_RESOURCE_GROUP_NAME` | Resource group to deploy to          | none                    |
| `KU_SENDGRID_APIKEY`     | Twilio SendGrid API key              | none                    |
| `KU_SENDGRID_FROM`       | Notification receiver E-mail address | none                    |
| `KU_SENDGRID_TO`         | Notification sender E-mail address   | none                    |
| `KU_SENDGRID_SUBJECT`    | Notification E-mail subject          | none                    |
| `KU_AKS_CLUSTER`         | AKS cluster resource name            | none                    |
| `KU_AKS_RESOURCE_GROUP`  | AKS cluster resource group           | none                    |
| `KU_LOCATION`            | Azure region to deploy to            | `westeurope`            |
| `KU_IMAGE`               | `kubeup` container image and tag     | `joergjo/kubeup:stable` |

### Quick and dirty AKS cluster deployment

If you don't have an existing AKS cluster, you can quickly create a single node cluster for testing using the Azure CLI:

```bash
$ az group create --name <aks-cluster-resource-group> \
    --location <region>
$ az aks create --resource-group <aks-cluster-resource-group> \
    --name <aks-cluster-name> \
    --location <region> \
    --node-count 1 \
    --generate-ssh-keys
```

## Building

Building `kubeup` requires [Go 1.18 or later](https://go.dev/dl/) on Windows, macOS or Linux. The command line examples shown below use bash syntax, but the commands also work in PowerShell or CMD.

```bash
$ cd kubeup
$ go build -o ./kubeup cmd/main.go
$ ./kubeup --help
```

The repo also contains task definitions and debug settings for Visual Studio Code.

Alternatively, you can use the included `Dockerfile` to build a Docker container image and run `kubeup`. A container image is also provided on [Docker Hub](https://hub.docker.com/repository/docker/joergjo/kubeup), so you don't need to build it yourself.

```bash
$ cd kubeup

$ # Build a local Docker image
$ docker compose build

$ # Run a container
$ docker-compose up -d

$ # Shut down the container
$ docker-compose down
```

You can override the container image's name and tag by exporting the environment variables `IMAGE` and `TAG` or adding them to an [`.env`](https://docs.docker.com/compose/environment-variables/#the-env-file) file.

## Running `kubeup`

By default, `kubeup` uses the `LogNotifier`. To use the `SendGridNotifier`, export the following environment variables before starting `kubeup`:

- `KU_SENDGRID_APIKEY`: Your SendGrid API key
- `KU_SENDGRID_FROM`: The E-mail's from address
- `KU_SENDGRID_FROM`: The E-mail's to address
- `KU_SENDGRID_SUBJECT`: The E-mail's subject

All four environment variables must be set to enable the `SendGridNotifier`.

You can either run `kubeup` with no arguments or pass the following arguments:

```bash
# Runs kubeup on its default port (8000) and default path (/webhook)
./kubeup

# Runs kubeup on a specific port (8765)
./kubeup -port 8765

# Runs kubeup on a specific path (/events)
./kubeup -path /events

```

## Testing

To manually test `kubeup`, you can use the included [sample request](testdata/sample.json) and any HTTP client like [curl](https://curl.se), [httpie](https://httpie.io), [wget](https://www.gnu.org/software/wget/), or [Postman](https://www.postman.com). Send the sample request to the `kubeup` endpoint using HTTP POST.

![Sample request in Postman](media/postman.png)
