#!/bin/bash
if [ -z "$KU_RESOURCE_GROUP_NAME" ]; then
    echo "KUBEUP_RESOURCE_GROUP_NAME is not set. Please set it to the name of the resource group to deploy to."
    exit 1
fi
if [ -z "$KU_SENDGRID_APIKEY" ]; then
    echo "KU_SENDGRID_APIKEY is not set. Please set it to your Twilio SendGrid API Key."
    exit 1
fi
if [ -z "$KU_SENDGRID_FROM" ]; then
    echo "KU_SENDGRID_FROM is not set. Please set it to sender E-mail adress."
    exit 1
fi
if [ -z "$KU_SENDGRID_TO" ]; then
    echo "KU_SENDGRID_TO is not set. Please set it to receiver E-mail adress."
    exit 1
fi
if [ -z "$KU_AKS_CLUSTER" ]; then
    echo "KU_AKS_CLUSTER is not set. Please set it to receiver E-mail adress."
    exit 1
fi
if [ -z "$KU_AKS_RESOURCE_GROUP" ]; then
    echo "KU_AKS_RESOURCE_GROUP is not set. Please set it to receiver E-mail adress."
    exit 1
fi

resource_group_name=$KU_RESOURCE_GROUP_NAME
location=${KU_LOCATION:-westeurope}
image=${KU_IMAGE:-joergjo/kubeup:stable}
timestamp=$(date +%s)

echo "Using resource group $resource_group_name in $location"

az group create \
  --resource-group "$resource_group_name" \
  --location "$location" \
  --output none

fqdn=$(az deployment group create \
  --resource-group "$resource_group_name" \
  --name "kubeup-webhook-$timestamp" \
  --template-file webhook.bicep \
  --parameters location="$location" sendGridApiKey="$KU_SENDGRID_APIKEY" \
    sendGridFrom="$KU_SENDGRID_FROM" sendGridTo="$KU_SENDGRID_TO" \
    sendGridSubject="$KU_SENDGRID_SUBJECT" image="$image" \
    appName="kubeup" \
  --query properties.outputs.fqdn.value \
  --output tsv)

az deployment group create \
  --resource-group "$KU_AKS_RESOURCE_GROUP" \
  --name "kubeup-eventgrid-$timestamp" \
  --template-file eventgrid.bicep \
  --parameters aksName="$KU_AKS_CLUSTER" \
    eventSubscriptionName="kubeup" \
    webhookUrl="https://$fqdn/webhook" \
  --output none

echo "Kubeup has been deployed successfully. The webhook URL is https://$fqdn"
