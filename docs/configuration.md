# Configuration Guide

## Creating an `.env` file for deployment
All tasks in the project's [Taskfile](../Taskfile.dist.yaml) rely on environment variables read from a `.env` file in the project's root directory. These variables are automatically loaded when tasks are executed, so you don't need to manage them manually in your shell.

To create your `.env` file, copy the included template and configure the variables for your deployment scenario. Remove or leave blank any variables you don't need.

```bash
cp .env.template .env
```

## Environment variables reference

| Variable                 | Purpose                                | Default             |
| ------------------------ | -------------------------------------- | ------------------- |
| `KU_PATH`                | Webhook URL path                       | `webhook`           |
| `KU_PORT`                | HTTP port                              | `8000`              |
| `KU_DEBUG`               | Enables debug-level logging            | `false`             |
| `KU_RESOURCE_GROUP`      | Azure resource group for deployment    | none                |
| `KU_LOCATION`            | Azure region                           | `westeurope`        |
| `KU_IMAGE`               | Container image and tag                | `joergjo/kubeup:latest` |
| `KU_AKS_CLUSTER`         | AKS cluster name                       | none                |
| `KU_AKS_RESOURCE_GROUP`  | AKS cluster resource group             | none                |
| `KU_EMAIL_FROM`          | Sender email address                   | none                |
| `KU_EMAIL_TO`            | Recipient email address                | none                |
| `KU_EMAIL_SUBJECT`       | Email subject                          | none                |
| `KU_SENDGRID_APIKEY`     | Twilio SendGrid API key                | none                |
| `KU_SMTP_HOST`           | SMTP hostname                          | none                |
| `KU_SMTP_PORT`           | SMTP port                              | `587`               |
| `KU_SMTP_USERNAME`       | SMTP username                          | none                |
| `KU_SMTP_PASSWORD`       | SMTP password                          | none                |
| `KU_SECRET_1`            | First client secret                    | none                |
| `KU_SECRET_2`            | Second client secret                   | none                |
| `KU_APP_ID`              | Entra ID application ID                | none                |

## Event delivery options
The following sections describe the environment variables needed for each `kubeup` feature.

### Logging
`kubeup` logs events at info level by default. To enable debug-level logging:

| Variable     | Value              | Default |
| ------------ | ------------------ | ------- |
| `KU_DEBUG`   | `true` or `false`  | `false` |

### Twilio SendGrid email delivery
To enable email delivery using Twilio SendGrid:

| Variable              | Value                   | Default |
| --------------------- | ----------------------- | ------- |
| `KU_SENDGRID_APIKEY`  | Twilio SendGrid API key | none    |
| `KU_EMAIL_FROM`       | Sender email address    | none    |
| `KU_EMAIL_TO`         | Recipient email address | none    |
| `KU_EMAIL_SUBJECT`    | Email subject           | none    |

### SMTP email delivery
To enable email delivery using SMTP:

| Variable           | Value                   | Default |
| ------------------ | ----------------------- | ------- |
| `KU_SMTP_HOST`     | SMTP hostname           | none    |
| `KU_SMTP_PORT`     | SMTP port               | `587`   |
| `KU_SMTP_USERNAME` | SMTP username           | none    |
| `KU_SMTP_PASSWORD` | SMTP password           | none    |
| `KU_EMAIL_FROM`    | Sender email address    | none    |
| `KU_EMAIL_TO`      | Recipient email address | none    |
| `KU_EMAIL_SUBJECT` | Email subject           | none    |

## AKS cluster monitoring
To specify the AKS cluster to monitor:

| Variable                | Value                  | Default |
| ----------------------- | ---------------------- | ------- |
| `KU_AKS_CLUSTER`        | Cluster name           | none    |
| `KU_AKS_RESOURCE_GROUP` | Cluster resource group | none    |

If these variables are not specified, the deployment will create the `kubeup` Azure Container App without an Azure Event Grid subscription. The webhook will run but receive no events. You can add these settings later and redeploy to create the Event Grid subscription.

## Webhook authorization
You can secure the webhook using Entra ID access tokens, client secrets, or disable authorization entirely. Note that without authorization, anyone who discovers your webhook URL can trigger it.

### No authorization
To disable authorization, leave these variables unset:

| Variable      | Value | Default |
| ------------- | ----- | ------- |
| `KU_SECRET_1` | none  | none    |
| `KU_SECRET_2` | none  | none    |
| `KU_APP_ID`   | none  | none    |

### Client secrets
Create two secure passwords and define:

| Variable      | Value              | Default |
| ------------- | ------------------ | ------- |
| `KU_SECRET_1` | Secure password #1 | none    |
| `KU_SECRET_2` | Secure password #2 | none    |
| `KU_APP_ID`   | none               | none    |

> **Tip:** Generate secure secrets with `openssl rand -hex 32`

### Entra ID
When using the deployment tasks with Entra ID authorization, `KU_APP_ID` is stored in `.env.entraid`. Both `.env` and `.env.entraid` are read by the deployment tasks.

For a manual setup, you can store the application ID in your `.env` file:

| Variable      | Value          | Default |
| ------------- | -------------- | ------- |
| `KU_SECRET_1` | none (ignored) | none    |
| `KU_SECRET_2` | none (ignored) | none    |
| `KU_APP_ID`   | Application ID | none    |

**Note:** Entra ID authorization takes precedence over client secrets. If both are configured, client secrets are ignored. 

## Command line arguments
While most configuration uses environment variables, `kubeup` supports a few command line arguments:

```bash
# Default port (8000) and path (/webhook)
./kubeup

# Custom path
./kubeup -path /events

# Custom port
./kubeup -port 8088

# Custom port and path
./kubeup -path /events -port 8088

# Enable debug logging
./kubeup -debug
```

Environment variables (i.e., `KU_PATH`, `KU_PORT`, `KU_DEBUG`) take precedence over their corresponding command line arguments.