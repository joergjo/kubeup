# Quickstart
This quickstart guides you to deploy `kubeup` on Azure securely using client secrets with no dependency on email services. All events received by `kubeup` will be logged to a [Log Analytics workspace](https://learn.microsoft.com/en-us/azure/container-apps/log-monitoring?source=recommendations&tabs=bash). 

## Prerequisites
To complete this quickstart, you need the following services and tools:

1. **Azure subscription** - Sign up [for free](https://azure.microsoft.com/free/)
2. **Bash environment** - Available on macOS/Linux by default, or [Windows Subsystem for Linux](https://docs.microsoft.com/en-us/windows/wsl/install) on Windows
3. **[Task](https://taskfile.dev)** - Build and deployment automation  
4. **[Azure CLI](https://docs.microsoft.com/cli/azure/install-azure-cli)** - Deploy Bicep templates to Azure
5. **[Bicep CLI](https://learn.microsoft.com/en-us/azure/azure-resource-manager/bicep/install#azure-cli)** - Install with `az bicep install`

Alternatively, use [Azure Cloud Shell](https://shell.azure.com) which has all required tools preinstalled.

`kubeup` is designed for Azure Kubernetes Service (AKS) cluster operators. If you already have an AKS cluster, skip to the [preparation of the `.env` file](#prepare-the-env-file).

## Deployment step-by-step
### Create an AKS cluster (optional)
If you don't have an AKS cluster, create a basic cluster for testing using the included [deployment script](../deploy/aks.sh). Adjust the cluster name, resource group, and region as needed:

```bash
cd ./deploy
export KU_AKS_CLUSTER='my-aks'
export KU_AKS_RESOURCE_GROUP='my-aks-rg'
export KU_LOCATION='northeurope'
./aks.sh
```

### Prepare the `.env` file
Create an `.env` file to store deployment configuration. Copy the template and edit it:

```bash
cp .env.template .env
```

Remove all entries from the template *except* those shown below. Provide your resource names, secrets, and preferred Azure region. If you created a cluster in the previous step, use the same names and region.

> **Tip:** Create secure client secrets with `openssl rand -hex 32`

```bash
# AKS cluster configuration
KU_AKS_CLUSTER='my-aks'
KU_AKS_RESOURCE_GROUP='my-aks-rg'

# kubeup deployment configuration
KU_APP_NAME='kubeup'
KU_RESOURCE_GROUP='kubeup-rg'
KU_LOCATION='northeurope'

# Client secrets for webhook authentication
KU_SECRET_1='...'
KU_SECRET_2='...'
```

## Deployment
Deploy `kubeup` to Azure using the included Bicep templates. This creates a `kubeup` Azure Container App and webhook subscription through Azure Event Grid for all events emitted by your AKS cluster:

```bash
task deploy:azure
```

### Verification 
After deployment, it may take time before receiving notifications, depending on Kubernetes version availability or cluster upgrades.

To test `kubeup` immediately, trigger a Kubernetes update to a newer Kubernetes version and check the logs:

```bash
version=$(az aks get-upgrades -n $KU_AKS_CLUSTER -g $KU_AKS_RESOURCE_GROUP \
  --query 'controlPlaneProfile.upgrades[].kubernetesVersion' -o tsv | sort | head -1)
az aks upgrade -n $KU_AKS_CLUSTER -g $KU_AKS_RESOURCE_GROUP -k $version -y
```

Once Kubernetes events are published, you'll see log entries in your Log Analytics workspace's `ContainerAppConsoleLogs_CL` table.

![Log entries in Azure Log Analytics](../media/logentry.png)

Query the last ten log entries with:
```
ContainerAppConsoleLogs_CL 
  | top 10 by time_d 
  | project Timestamp=unixtime_seconds_todatetime(time_d), Log_data_s, Log_msg_s 
````

### Next steps
See the [full deployment documentation](./deployment.md) to learn how to protect `kubeup` using Entra ID and enable email notifications.