@description('Specifies the location to deploy to.')
param location string = resourceGroup().location

@description('Specifies the Container App\'s name.')
@minLength(5)
@maxLength(12)
param appName string = 'kubeup'

@description('Specifies the Container App\'s image.')
param image string

@description('Specifies the notification\'s email From address.')
param emailFrom string

@description('Specifies the notification\'s email To address.')
param emailTo string

@description('Specifies the notification\'s email subject´.')
param emailSubject string

@description('Specifies the Twilio SendGrid API Key.')
@secure()
param sendGridApiKey string

@description('Specifies the SMTP hostname.')
param smtpHost string

@description('Specifies the SMTP port.')
param smtpPort int = 587

@description('Specifies the SMTP username.')
@secure()
param smtpUsername string

@description('Specifies the SMTP password.')
@secure()
param smtpPassword string

var namePrefix = '${appName}-${uniqueString(resourceGroup().id)}'

module network 'modules/network.bicep' = {
  name: 'network'
  params: {
    location: location
    namePrefix: namePrefix
  }
}

module environment 'modules/environment.bicep' = {
  name: 'environment'
  params: {
    location: location
    namePrefix: namePrefix
    infrastructureSubnetId: network.outputs.infraSubnetId
  }
}

module app 'modules/app.bicep' = {
  name: 'app'
  params: {
    name: appName
    location: location
    environmentId: environment.outputs.environmentId
    image: image
    emailFrom: emailFrom
    emailTo: emailTo
    emailSubject: emailSubject
    sendGridApiKey: sendGridApiKey
    smtpHost: smtpHost
    smtpPort: smtpPort
    smtpUsername: smtpUsername
    smtpPassword: smtpPassword
  }
}

output fqdn string = app.outputs.fqdn
