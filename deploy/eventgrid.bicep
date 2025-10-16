@description('Specifies the AKS cluster name.')
param aksName string

@description('Specifies the webhook URL to deliver events to.')
@secure()
param webhookUrl string

@description('Specifies the webhook\'s Entra ID application ID or URI.')
param appId string

var enableEntraId = appId != '' 

resource aks 'Microsoft.ContainerService/managedClusters@2025-07-01' existing = {
  name: aksName
}

resource eventSubscription 'Microsoft.EventGrid/eventSubscriptions@2025-02-15' = {
  name: 'kubeup-${uniqueString(webhookUrl, resourceGroup().id)}'
  scope: aks
  properties: {
    destination: {
      endpointType: 'WebHook'
      properties: enableEntraId ? {
        endpointUrl: webhookUrl
        minimumTlsVersionAllowed: '1.2'
        azureActiveDirectoryApplicationIdOrUri: appId
#disable-next-line use-resource-id-functions
        azureActiveDirectoryTenantId: tenant().tenantId
      } : {
        endpointUrl: webhookUrl
        minimumTlsVersionAllowed: '1.2'
      }
    }
    eventDeliverySchema: 'CloudEventSchemaV1_0'
    filter: {
      includedEventTypes: [
        'Microsoft.ContainerService.NewKubernetesVersionAvailable'
        'Microsoft.ContainerService.ClusterSupportEnded'
        'Microsoft.ContainerService.ClusterSupportEnding'
        'Microsoft.ContainerService.NodePoolRollingFailed'
        'Microsoft.ContainerService.NodePoolRollingStarted'
        'Microsoft.ContainerService.NodePoolRollingSucceeded'
      ]
    }
  }
}
