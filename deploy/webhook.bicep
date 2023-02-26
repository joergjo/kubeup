@description('Specifies the location to deploy to.')
param location string = resourceGroup().location

@description('Specifies the Container App\'s name.')
@minLength(5)
@maxLength(12)
param appName string

@description('Specifies the Container App\'s image.')
param image string

@description('Specifies the Twilio SendGrid API Key.')
@secure()
param sendGridApiKey string

@description('Specifies the Twilio SendGrid E-mail from address.')
param sendGridFrom string

@description('Specifies the Twilio SendGrid E-mail to address.')
param sendGridTo string

@description('Specifies the Twilio SendGrid E-mail subject.')
param sendGridSubject string

@description('Specifies the SMTP hostname.')
param smptHost string

@description('Specifies the SMTP port.')
param smptPort int = 587

@description('Specifies the SMTP username.')
@secure()
param smptUsername string

@description('Specifies the SMTP password.')
@secure()
param smptPassword string

@description('Specifies the SMTP from address.')
param smptFrom string

@description('Specifies the SMTP to address.')
param smptTo string

@description('Specifies the SMTP subject.')
param smptSubject string

module network 'modules/network.bicep' = {
  name: 'network'
  params: {
    location: location
    vnetName: '${appName}-vnet'
  }
}

module environment 'modules/environment.bicep' = {
  name: 'environment'
  params: {
    location: location
    namePrefix: appName
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
    sendGridApiKey: sendGridApiKey
    sendGridFrom: sendGridFrom
    sendGridTo: sendGridTo
    sendGridSubject: sendGridSubject
    smptHost: smptHost
    smptPort: smptPort
    smptUsername: smptUsername
    smptPassword: smptPassword
    smptFrom: smptFrom
    smptTo: smptTo
    smptSubject: smptSubject
  }
}

output fqdn string = app.outputs.fqdn
