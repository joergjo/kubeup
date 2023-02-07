@description('Specifies the name of the Container App.')
param name string

@description('Specifies the location to deploy to.')
param location string

@description('Specifies the name of Azure Container Apps environment to deploy to.')
param environmentId string

@description('Specifies the container image.')
param image string

@description('Specifies the SendGrid API key.')
@secure()
param sendGridApiKey string

@description('Specifies the notification\'s email From address.')
param sendGridFrom string

@description('Specifies the notification\'s email To address.')
param sendGridTo string

@description('Specifies the notification\'s email subject´.')
param sendGridSubject string

var port = 8000

resource containerApp 'Microsoft.App/containerApps@2022-03-01' = {
  name: name
  location: location
  properties: {
    managedEnvironmentId: environmentId
    configuration: {
      secrets: [
        {
          name: 'sendgrid-api-key'
          value: sendGridApiKey
        }
      ]
      ingress: {
        external: true
        targetPort: port
      }
      dapr: {
        enabled: false
      }
    }
    template: {
      containers: [
        {
          image: image
          name: name
          env: [
            {
              name: 'KU_SENDGRID_API_KEY'
              secretRef: 'sendgrid-api-key'
            }
            {
              name: 'KU_SENDGRID_FROM'
              value: sendGridFrom
            }
            {
              name: 'KU_SENDGRID_TO'
              value: sendGridTo
            }
            {
              name: 'KU_SENDGRID_SUBJECT'
              value: sendGridSubject
            }
          ]
          resources: {
            cpu: json('0.5')
            memory: '1Gi'
          }
          probes: [
            {
              type: 'liveness'
              httpGet: {
                path: '/healthz'
                port: port
              }
            }
            {
              type: 'readiness'
              httpGet: {
                path: '/healthz'
                port: port
              }
            }
          ]
        }
      ]
      scale: {
        minReplicas: 0
        maxReplicas: 2
        rules: [
          {
            name: 'httpscale'
            http: {
              metadata: {
                concurrentRequests: '100'
              }
            }
          }
        ]
      }
    }
  }
}

output fqdn string = containerApp.properties.configuration.ingress.fqdn
