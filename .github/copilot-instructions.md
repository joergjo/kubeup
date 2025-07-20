This is a Go based repository that provides `kubeup`, a webhook for handling Azure Kubernetes Service (AKS) [CloudEvents](https://cloudevents.io), providing real-time notifications on version updates and upgrade events emitted by AKS clusters. Rather than manually checking for new Kubernetes versions or monitoring upgrade completion, `kubeup` keeps you informed by automatically sending notifications whenever an update or upgrade event occurs. 

## Code Standards

### Required Before Each Commit
- Run `task go:tidy` before committing any changes to ensure proper code formatting

### Development Flow
- Build: `task go:build`
- Test: `task go:test`

## Repository Structure
The repository has the followingg structure:
- cmd/webhook: Contains the main application code for the `kubeup` webhook.
- internal/**: Contains internal packages that implement the webhook logic, including authorization, CloudEvents processing, and event handling.
- testdata: Contains test data used in unit tests and for manual testing.
- deploy: Contains deployment scripts and Bicep templates for deploying `kubeup` to Azure.
- docs: Contains documentation for deployment, configuration, and architecture.
- media: Contains media files, such as diagrams and images used in documentation.
- .github: Contains GitHub Actions workflows for building and testing the application.

The project's root folder contains the following important files: 
- `Taskfile.dist.yaml`, `Taskfile.Go.yaml` and `Taskfile.Docker.yaml`: These define tasks for building, testing, and deploying the application. The Taskfile is used to automate common development tasks and ensure consistency across the development team. - `Dockerfile` for building the `kubeup` container image.
- `compose.yaml` file for running the application in a Docker Compose environment.
- `.env.template`: A template for the environment variables required for deployment. This file should be copied to `.env` and modified with the appropriate values before deploying `kubeup`.

## Key Guidelines
1. Follow Go best practices and idiomatic patterns
2. Maintain existing code structure and organization
3. Write unit tests for new functionality. Use table-driven unit tests when possible.
