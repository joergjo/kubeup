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

@description('Specifies the SMTP hostname.')
param smtpHost string

@description('Specifies the SMTP port.')
param smtpPort int

@description('Specifies the SMTP username.')
@secure()
param smtpUsername string

@description('Specifies the SMTP password.')
@secure()
param smtpPassword string

@description('Specifies the SMTP from address.')
param smtpFrom string

@description('Specifies the SMTP to address.')
param smtpTo string

@description('Specifies the SMTP subject.')
param smtpSubject string

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
]

var secrets = filter(allSecrets, s => !empty(s.value))

var allEnvVars = [
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
    name: 'KU_SMTP_FROM'
    value: smtpFrom
  }
  {
    name: 'KU_SMTP_TO'
    value: smtpTo
  }
  {
    name: 'KU_SMTP_SUBJECT'
    value: smtpSubject
  }
]

var secretNames = map(secrets, s => s.name)

var envVars = filter(allEnvVars, e => (contains(e, 'secretRef') && contains(secretNames, e.secretRef)) || contains(e, 'value'))

resource containerApp 'Microsoft.App/containerApps@2022-03-01' = {
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
  }
}

output fqdn string = containerApp.properties.configuration.ingress.fqdn
