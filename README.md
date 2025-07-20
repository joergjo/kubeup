# kubeup
`kubeup` is a webhook for handling Azure Kubernetes Service (AKS) [CloudEvents](https://cloudevents.io), providing real-time notifications on version updates and upgrade events emitted by AKS clusters. Rather than manually checking for new Kubernetes versions or monitoring upgrade completion, `kubeup` keeps you informed by automatically sending notifications whenever an update or upgrade event occurs.

To learn more about the types of events AKS can emit and the different subscription methods, refer to [Quickstart: Subscribe to Azure Kubernetes Service (AKS) events with Azure Event Grid](https://docs.microsoft.com/en-us/azure/aks/quickstart-event-grid) and [Webhook event delivery](https://docs.microsoft.com/en-us/azure/event-grid/webhook-event-delivery). `kubeup` supports all [AKS event types](https://learn.microsoft.com/en-us/azure/event-grid/event-schema-aks?tabs=cloud-event-schema#available-event-types). 

`kubeup` can handle these events in several ways:
- Log events by writing them to standard error (stderr).
- Send notifications via email using SMTP.
- Send email notifications through Twilio SendGrid.

You can also extend kubeup to support other notification channels and delivery methods.

>AKS has a built-in feature called [AKS Communication Manager](https://learn.microsoft.com/en-us/azure/aks/aks-communication-manager), which provides alerts for planned maintenance of AKS clusters. AKS Communication Manager and `kubeup` are complementary and can be used together.

## Quickstart
If you want to deploy `kubeup` right away, follow the [quickstart instructions](./docs/quickstart.md). This repo includes Bicep templates to deploy `kubeup` as an [Azure Container App](https://docs.microsoft.com/en-us/azure/container-apps/overview), including [HTTP scaling rules to scale to zero](https://docs.microsoft.com/en-us/azure/container-apps/scale-app). This ensures that `kubeup` can be run with minimal costs.

## Deployment guide
If you want to learn about all deployment and configuration options, see the [deployment guide](./docs/deployment.md).

## Additional resources
The [implementation documentation](./docs/implementation.md) provides details on `kubeup`'s implementation. 

If you want to dive into the source code and build your own version of `kubeup`, see the [developer documentation](./docs/development.md). 
