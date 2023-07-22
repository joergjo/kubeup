@description('Specifies the name of the virtual network.')
param vnetName string

@description('Specifies the location to deploy to.')
param location string

resource vnet 'Microsoft.Network/virtualNetworks@2022-09-01' = {
  name: vnetName
  location: location
  properties: {
    addressSpace: {
      addressPrefixes: [
        '10.150.0.0/22'
      ]
    }
    subnets: [
      {
        name: 'infrastructure'
        properties: {
          addressPrefix: '10.150.0.0/23'
          delegations: [
            {
              name: 'Microsoft.App/environments'
              properties: {
                serviceName: 'Microsoft.App/environments'
              }
            }
          ]
        }
      }
    ]
  }
}

output vnetId string = vnet.id
output infraSubnetId string = vnet.properties.subnets[0].id
