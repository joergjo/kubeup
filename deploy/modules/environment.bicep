@description('Specifies the name prefix of all resources.')
param namePrefix string

@description('Specifies the location to deploy to.')
param location string

@description('Specifies the subnet resource ID for the Container App environment.')
param infrastructureSubnetId string

var workspaceName = '${namePrefix}-logs'
var environmentName = '${namePrefix}-env'

resource logAnalyticsWorkspace 'Microsoft.OperationalInsights/workspaces@2025-02-01' = {
  name: workspaceName
  location: location
  properties: {
    sku: {
      name: 'PerGB2018'
    }
  }
}

resource environment 'Microsoft.App/managedEnvironments@2026-07-01' = {
  name: environmentName
  location: location
  properties: {
    appLogsConfiguration: {
      destination: 'azure-monitor'
    }
    vnetConfiguration: {
      infrastructureSubnetId: infrastructureSubnetId
    }
    workloadProfiles: [
      {
        name: 'Consumption'
        workloadProfileType: 'Consumption'
      }
    ]
  }
}

resource diagnostics 'Microsoft.Insights/diagnosticSettings@2021-05-01-preview' = {
  name: '${environmentName}-logs'
  scope: environment
  properties: {
    workspaceId: logAnalyticsWorkspace.id
    logs: [
      {
        category: 'ContainerAppSystemLogs'
        enabled: true
      }
      {
        category: 'ContainerAppConsoleLogs'
        enabled: true
      }
      {
        category: 'ContainerAppHTTPLogs'
        enabled: true
      }
    ]
  }
}

output environmentId string = environment.id
