@description('Specifies the Event Grid subscription name.')
param eventSubscriptionName string

@description('Specifies the AKS cluster name.')
param aksName string

@description('Specifies the webhook URL to deliver events to.')
param webhookUrl string

resource aks 'Microsoft.ContainerService/managedClusters@2023-02-01' existing = {
  name: aksName
}

resource eventSubscription 'Microsoft.EventGrid/eventSubscriptions@2022-06-15' = {
  name: eventSubscriptionName
  scope: aks
  properties: {
    destination: {
      endpointType: 'WebHook'
      properties: {
        endpointUrl: webhookUrl
      }
    }
    eventDeliverySchema: 'CloudEventSchemaV1_0'
    filter: {
      includedEventTypes: [
        'Microsoft.ContainerService.NewKubernetesVersionAvailable'
      ]
    }
  }
}
