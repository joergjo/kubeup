# kubeup

`kubeup` is a sample WebHook written in [Go](https://go.dev) to process Azure Kubernetes Service (AKS) [CloudEvents](https://cloudevents.io) that notify receivers of new Kubernetes versions being available in AKS. Refer to [Quickstart: Subscribe to Azure Kubernetes Service (AKS) events with Azure Event Grid (Preview)](https://docs.microsoft.com/en-us/azure/aks/quickstart-event-grid) and [WebHook Event delivery](https://docs.microsoft.com/en-us/azure/event-grid/webhook-event-delivery) if you want to learn more about the underlying concepts.

Events received by `kubeup` are handled internally by a `Publisher`, which is a struct that holds a slice of `PublisherFunc` functions. `kubeup` provides various implemnetations out of the box:

- Write to stderr using [zerolog](github.com/rs/zerolog).
- Send an email using the [Twilio SendGrid](https://sendgrid.com) API.
- Send an email using SMTP.
- Provide your own `PublisherFunc`.

`kubeup` does _not_ implement any authorization (yet). For a production grade implemetation, you should [secure your WebHook endpoint with Azure AD](https://docs.microsoft.com/en-us/azure/event-grid/secure-webhook-delivery).

Since Azure Event Grid delivers events only to public endpoints, you must either run `kubeup` on an Azure service that allows you to expose a public endpoint (App Service, Container App, AKS, VMs, etc.), or use a reverse proxy service like [ngrok](https://ngrok.com) to route events to a local endpoint. This repo includes Bicep templates to deploy `kubeup` as an [Azure Container App](https://docs.microsoft.com/en-us/azure/container-apps/overview), including [HTTP scaling rules to scale to zero](https://docs.microsoft.com/en-us/azure/container-apps/scale-app).

## Quickstart

### Prerequisites

1. An Azure subscription. Sign up [for free](https://azure.microsoft.com/free/).
2. Access to the [Azure CLI](https://docs.microsoft.com/cli/azure/install-azure-cli), either installed locally or using the [Azure Cloud Shell](https://shell.azure.com). Make sure to have Bicep CLI installed as well by running `az bicep install`.
3. A bash shell to execute the included deployment script - on Windows 10/11 use the [Window Subsystem for Linux](https://docs.microsoft.com/en-us/windows/wsl/install).
4. If you want to send notifications through email, you need either a [Twilio SendGrid account](https://sendgrid.com/pricing/) or access to an SMTP host to send email. [Mailtrap](https://mailtrap.io) has a free tier that works great for this. 

### Creating an AKS cluster

If you don't have an existing AKS cluster, you can create a small cluster to test `kubeup` using the Azure CLI:

```bash
$ az group create --name <aks-cluster-resource-group> \
    --location <region>
$ az aks create --resource-group <aks-cluster-resource-group> \
    --name <aks-cluster-name> \
    --location <region> \
    --node-count 2 \
    --generate-ssh-keys
```

Just keep in mind that it may take some time before you will receive a notification. 

### Deployment

Use the included deployment script to deploy `kubeup` to an Azure Container App that uses logging to stderr, Twilio SendGrid, or SMTP depending on the configuration you provide. The Bicep templates will both deploy a `kubeup` Azure Container App and create a WebHook subscription for your AKC cluster.

```bash
$ export KU_RESOURCE_GROUP=my-kubeup-rg
$ cd kubeup/deploy
$ ./deploy.sh
```

All resources are created in the same region. You can override the default settings
of the deployment script by exporting the following environment variables. Note that all
environment variables with no default value are required.

| Environment variable     | Purpose                                | Default value           |
| ------------------------ | ---------------------------------------| ----------------------- |
| `KU_RESOURCE_GROUP`      | Resource group to deploy to            | none                    |
| `KU_LOCATION`            | Azure region to deploy to              | `westeurope`            |
| `KU_IMAGE`               | `kubeup` container image and tag       | `joergjo/kubeup:stable` |
| `KU_AKS_CLUSTER`         | AKS cluster resource name              | none                    |
| `KU_AKS_RESOURCE_GROUP`  | AKS cluster resource group             | none                    |
| `KU_SENDGRID_APIKEY`     | Twilio SendGrid API key                | none                    |
| `KU_SENDGRID_FROM`       | Twilio SendGrid receiver email address | none                    |
| `KU_SENDGRID_TO`         | Twilio SendGrid sender email address   | none                    |
| `KU_SENDGRID_SUBJECT`    | Twilio SendGrid email subject          | none                    |
| `KU_SMTP_HOST`           | SMTP hostname                          | none                    |
| `KU_SMTP_PORT`           | SMTP port                              | `587`                    |
| `KU_SMTP_USERNAME`       | SMTP username                          | none                    |
| `KU_SMTP_PASSWORD`       | SMTP password                          | none                    |
| `KU_SMTP_FROM`           | SMTP sender email address              | none                    |
| `KU_SMTP_TO`             | SMTP receiver email address            | none                    |
| `KU_SMTP_SUBJECT`        | SMTP email subject                     | none                    |

If you do not provide `KU_AKS_CLUSTER` and `KU_AKS_RESOURCE_GROUP`, the script will only deploy
the `kubeup` webhook. You can rerun the deployment script later again with `KU_AKS_CLUSTER` and `KU_AKS_RESOURCE_GROUP`set to complete the deployment.

Now, once Kubernetes upgrades are published for your AKS cluster, you will receive an email (if configured) and find a log entry in your Log Analytics workspace's `ContainerAppConsoleLogs_CL` table.

## Building `kubeup`

Building `kubeup` requires [Go 1.20 or later](https://go.dev/dl/) on Windows, macOS or Linux. The command line examples shown below use bash syntax, but the commands also work in PowerShell or CMD on Windows by substituting `/` with `\`.

```bash
$ cd kubeup
$ go test -v ./...
$ go build -o ./kubeup ./cmd/main.go
$ ./kubeup --help
```

The repo also contains task definitions and debug settings for Visual Studio Code.

## Docker support

`kubeup` container images for both AMD64 and ARM64 architectures are available on [Docker Hub](https://hub.docker.com/repository/docker/joergjo/kubeup). 
You can use the included `Dockerfile` to build your own container image instead and run `kubeup` in Docker, Podman etc. 

```bash
$ cd kubeup

$ # Build a local Docker image
$ docker compose build

$ # Run a container
$ docker compose up -d

$ # Shut down the container
$ docker compose down
```

You can override the container image's name and tag by exporting the environment variables `IMAGE` and `TAG` or adding them to an [`.env`](https://docs.docker.com/compose/environment-variables/#the-env-file) file.

## Running `kubeup`

Out of the box, `kubeup` writes all notifications to stderr. It supports the following arguments:

```bash
# Runs kubeup on its default port (8000) and default path (/webhook)
./kubeup

# Runs kubeup on a specific path (/events)
./kubeup -path /events

# Runs kubeup on a specific port (:8088)
./kubeup -port 8088

# Runs kubeup on a specific port and path (:8088/events)
./kubeup -path /events -port 8088

```

### Twilio SendGrid email delivery

To enable email delivery using Twilio SendGrid, export the following environment variables before starting `kubeup`:

| Environment variable     | Purpose                                | Default value           |
| ------------------------ | ---------------------------------------| ----------------------- |
| `KU_SENDGRID_APIKEY`     | Twilio SendGrid API key                | none                    |
| `KU_SENDGRID_FROM`       | Twilio SendGrid receiver email address | none                    |
| `KU_SENDGRID_TO`         | Twilio SendGrid sender email address   | none                    |
| `KU_SENDGRID_SUBJECT`    | Twilio SendGrid email subject          | none                    |

All environment variables must be exported.

### SMTP email delivery

To enable email delivery using SMTP, export the following environment variables before starting `kubeup`:

| Environment variable     | Purpose                                | Default value           |
| ------------------------ | ---------------------------------------| ----------------------- |
| `KU_SMTP_HOST`           | SMTP hostname                          | none                    |
| `KU_SMTP_PORT`           | SMTP port                              | `587`                    |
| `KU_SMTP_USERNAME`       | SMTP username                          | none                    |
| `KU_SMTP_PASSWORD`       | SMTP password                          | none                    |
| `KU_SMTP_FROM`           | SMTP sender email address              | none                    |
| `KU_SMTP_TO`             | SMTP receiver email address            | none                    |
| `KU_SMTP_SUBJECT`        | SMTP email subject                     | none                    |

All environment variables must be exported except `KU_SMTP_PORT`.

## Testing

To manually test `kubeup`, you can use the included [sample request](testdata/sample.json) and any HTTP client like [curl](https://curl.se), [httpie](https://httpie.io), [wget](https://www.gnu.org/software/wget/), or [Postman](https://www.postman.com). Send the sample request to the `kubeup` endpoint using HTTP POST.

![Sample request in Postman](media/postman.png)
