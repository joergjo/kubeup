#!/bin/bash
if [ -z "$KU_RESOURCE_GROUP" ]; then
    echo "KU_RESOURCE_GROUP is not set. Please set it to the name of the resource group to deploy to."
    exit 1
fi

resource_group=$KU_RESOURCE_GROUP
location=${KU_LOCATION:-westeurope}
image=${KU_IMAGE:-joergjo/kubeup:latest}
timestamp=$(date +%s)

echo "Using resource group $resource_group in $location"

az group create \
  --resource-group "$resource_group" \
  --location "$location" \
  --output none

fqdn=$(az deployment group create \
  --resource-group "$resource_group" \
  --name "kubeup-webhook-$timestamp" \
  --template-file webhook.bicep \
  --parameters location="$location" image="$image" appName="kubeup" \
    sendGridApiKey="$KU_SENDGRID_APIKEY" sendGridFrom="$KU_SENDGRID_FROM" \
    sendGridTo="$KU_SENDGRID_TO" sendGridSubject="$KU_SENDGRID_SUBJECT" \
    smtpHost="$KU_SMTP_HOST" smtpPort="$KU_SMTP_PORT" \
    smtpUsername="$KU_SMTP_USERNAME" smtpPassword="$KU_SMTP_PASSWORD" \
    smtpFrom="$KU_SMTP_FROM" smtpTo="$KU_SMTP_TO" \
    smtpSubject="$KU_SMTP_SUBJECT" \
  --query properties.outputs.fqdn.value \
  --output tsv)

if [ -z "$fqdn" ]; then
    echo "Failed to create webhook."
    exit 1
fi

echo "Kubeup has been deployed successfully. The webhook URL is https://$fqdn"

if [ -z "$KU_AKS_CLUSTER" || -z "$KU_AKS_RESOURCE_GROUP" ]; then
    echo "KU_AKS_CLUSTER or KU_AKS_RESOURCE_GROUP not set. Skipping Event Grid topic creation."
    exit 0
fi

az deployment group create \
  --resource-group "$KU_AKS_RESOURCE_GROUP" \
  --name "kubeup-eventgrid-$timestamp" \
  --template-file eventgrid.bicep \
  --parameters aksName="$KU_AKS_CLUSTER" \
    eventSubscriptionName="kubeup" \
    webhookUrl="https://$fqdn/webhook" \
  --output none

echo "Event Grid topic has been created successfully."
