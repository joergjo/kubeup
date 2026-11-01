# Development

The included Taskfiles contain task definitions to build, test, run, and deploy `kubeup` without requiring knowledge of Go, Docker, or Azure CLI toolchains. You can also use Visual Studio Code to build and debug `kubeup`.

## Prerequisites
1. **[Go](https://go.dev/dl/) 1.24 or later** - Go toolchain and standard library
2. **[Task](https://taskfile.dev)** - Build and deployment automation
3. **[Azure CLI](https://docs.microsoft.com/cli/azure/install-azure-cli)** - Deploy Bicep templates to Azure
4. **[Bicep CLI](https://learn.microsoft.com/azure/azure-resource-manager/bicep/install#azure-cli)** - Install with `az bicep install`
5. **[Visual Studio Code](https://code.visualstudio.com/)** (optional) - Configured IDE with debugging support

## Terminal-based development
Use the included Taskfiles to test, build, run, and deploy without knowing the underlying toolchains.

### Run kubeup
Compiles, runs, and cleans up the binary. Settings from local `.env` files are applied automatically.

```bash
task go:run
# or simply
task
```

### Run tests
Executes application/package Go tests. These tests do not read deployment files or require Bash, jq, Azure CLI, or other deployment tools:

```bash
task go:test
```

Deployment checks are separate:

| Command | Purpose | Additional tools |
| --- | --- | --- |
| `task deploy:test` | Bats tests of deployment scripts with isolated CLI mocks | Bash, Bats-core, jq |
| `task deploy:contract` | Explicit Go contract check between Bicep variables and application configuration tags | Go |
| `task deploy:lint` | ShellCheck and offline Bicep build/lint | ShellCheck, Azure CLI, Bicep |
| `task check` | All application and deployment checks | All of the above |

Install [Bats-core](https://github.com/bats-core/bats-core), [ShellCheck](https://www.shellcheck.net/), and jq with `brew install bats-core shellcheck jq` on macOS or `sudo apt-get install bats shellcheck jq` on Debian/Ubuntu. Install Bicep with `az bicep install`. Windows users should run deployment checks in WSL or another Bash environment. Check tasks report missing tools rather than installing or silently skipping them.

Script tests live in `deploy/tests`; they use temporary workspaces and mock `az`, `uuidgen`, and `sleep`. They do not call Azure or use developer credentials. The environment-variable contract lives in `deploy/tests/contract` behind the `deployment_contract` build tag, so ordinary `go test ./...` does not run it. It parses application configuration tags without importing or exporting application internals.

GitHub Actions runs application tests and deployment checks in separate jobs on pull requests and pushes to `main`, without cloud credentials. Deployment jobs install pinned tool versions; local package-manager versions may differ.

### Build executable
Compiles source code to a platform-specific executable:

```bash
task go:build
```

### Format and tidy
Formats code per Go standards and cleans up dependencies:

```bash
task go:tidy
```

### Build container image
Builds a container image from current source code and caches it locally (not pushed to registry).

Override image name/tag with `IMAGE` and `TAG` environment variables or in your `.env` file. Defaults to `kubeup:dev`.

```bash
task docker:build
```

### Run container image
Requires `task docker:build` first. Runs container and exposes webhook on host port 8000.

```bash 
task docker:up
```

### Stream container logs
Requires `task docker:up` first. Streams running container logs to terminal.

```bash
task docker:logs 
```

### Stop container
Shuts down the running container.

```bash
task docker:down 
```

## Visual Studio Code development
Open the project folder in VS Code. It will prompt to install recommended extensions if not already installed.

### Building
Execute **Run Build Task** (Cmd+Shift+B / Ctrl+Shift+B).

![Building in Visual Studio Code](../media/vscode-build.png)

### Debugging
Execute **Debug: Start Debugging** (F5). Requires a valid `.env` file in the project root. See [deployment instructions](deployment.md) to create one.

![Debugging in Visual Studio Code](../media/vscode-debug.png)

### Running Task tasks
With the [Task extension](https://marketplace.visualstudio.com/items?itemName=task.vscode-task) installed, run tasks directly from the Task sidebar.

![Running Task tasks in Visual Studio Code](../media/vscode-task.png)

## Manual testing
Test `kubeup` using the included [sample requests](../testdata/) and any HTTP client like curl, httpie, wget, or Postman. Send sample requests to the webhook endpoint via HTTP POST.

![Sample request in Postman](../media/postman.png)

## Customizing task definitions
To add or modify task definitions, create `Taskfile.yaml` in the project root and adapt content from `Taskfile.dist.yaml`. This personal task file [takes precedence](https://taskfile.dev/usage/) and is ignored by Git.  
