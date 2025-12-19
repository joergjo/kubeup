@description('Specifies the name of the Container App.')
param name string

@description('Specifies the location to deploy to.')
param location string

@description('Specifies the name of Azure Container Apps environment to deploy to.')
param environmentId string

@description('Specifies the container image.')
param image string

@description('Specifies the primary client secret.')
@secure()
param secret1 string

@description('Specifies the secondary client secret.')
@secure()
param secret2 string

@description('Specifies the webhook\'s Entra ID application ID or URI.')
param appId string

@description('Specifies the notification\'s email From address.')
param emailFrom string

@description('Specifies the notification\'s email To address.')
param emailTo string

@description('Specifies the notification\'s email subject´.')
param emailSubject string

@description('Specifies the SendGrid API key.')
@secure()
param sendGridApiKey string

@description('Specifies the SMTP hostname.')
param smtpHost string

@description('Specifies the SMTP port.')
param smtpPort string

@description('Specifies the SMTP username.')
@secure()
param smtpUsername string

@description('Specifies the SMTP password.')
@secure()
param smtpPassword string

var tenantId = appId != '' ? tenant().tenantId : ''
var port = 8000

var allSecrets = [
  {
    name: 'sendgrid-api-key'
    value: sendGridApiKey
  }
  {
    name: 'smtp-username'
    value: smtpUsername
  }
  {
    name: 'smtp-password'
    value: smtpPassword
  }
  {
    name: 'secret-1'
    value: secret1
  }
  {
    name: 'secret-2'
    value: secret2
  }
]

var secrets = filter(allSecrets, s => !empty(s.value))

var allEnvVars = [
  {
    name: 'KU_EMAIL_FROM'
    value: emailFrom
  }
  {
    name: 'KU_EMAIL_TO'
    value: emailTo
  }
  {
    name: 'KU_EMAIL_SUBJECT'
    value: emailSubject
  }
  {
    name: 'KU_SENDGRID_API_KEY'
    secretRef: 'sendgrid-api-key'
  }
  {
    name: 'KU_SMTP_HOST'
    value: smtpHost
  }
  {
    name: 'KU_SMTP_PORT'
    value: string(smtpPort)
  }
  {
    name: 'KU_SMTP_USERNAME'
    secretRef: 'smtp-username'
  }
  {
    name: 'KU_SMTP_PASSWORD'
    secretRef: 'smtp-password'
  }
  {
    name: 'KU_SECRET_1'
    secretRef: 'secret-1'
  }
  {
    name: 'KU_SECRET_2'
    secretRef: 'secret-2'
  }
  {
    name: 'KU_TENANT_ID'
    value: tenantId
  }
  {
    name: 'KU_APP_ID'
    value: appId
  }
]

var secretNames = map(secrets, s => s.name)

var envVars = filter(
  allEnvVars,
  e =>
    (contains(e, 'secretRef') && contains(secretNames, any(e).secretRef)) || contains(e, 'value') && !empty(any(e).value)
)

resource containerApp 'Microsoft.App/containerApps@2025-10-02-preview' = {
  name: name
  location: location
  properties: {
    managedEnvironmentId: environmentId
    configuration: {
      secrets: secrets
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
          env: envVars
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
    workloadProfileName: 'Consumption'
  }
}

output fqdn string = containerApp.properties.configuration.ingress.fqdn
