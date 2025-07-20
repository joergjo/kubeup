#!/bin/bash
echo "Ensuring resource group $KU_AKS_RESOURCE_GROUP exists..."
az group create --name $KU_AKS_RESOURCE_GROUP \
    --location $KU_LOCATION \
    --output none
echo "Creating AKS cluster $KU_AKS_CLUSTER..."
az aks create --resource-group $KU_AKS_RESOURCE_GROUP \
    --name $KU_AKS_CLUSTER \
    --location $KU_LOCATION \
    --node-count 2 \
    --generate-ssh-keys \
    --output none
echo "Successfully created AKS cluster $KU_AKS_CLUSTER"