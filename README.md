# Temporal Go Demo Application

A minimal Temporal application demonstrating workflow orchestration with a simple Hello World example. This project shows how to set up a basic Temporal workflow that greets a user.

## Prerequisites

Before you begin, ensure you have:
* Go 1.16+ installed
* [Temporal CLI](https://github.com/temporalio/cli) installed

## Project Structure

```
.
├── config/
│   └── temporal_config.go         # Temporal client configuration
├── workflows/
│   ├── workflow.go               # Workflow interface definition
│   └── impl/
│       └── hello_world_workflow.go # Workflow implementation
├── workers/
│   └── hello_world_worker.go     # Worker process to execute workflows
├── starter/
│   └── hello_world_starter.go    # Workflow starter implementation
└── main.go                       # Application entry point
```

## Configuration Options

The application supports both local development and Temporal Cloud deployments:

### Local Development (Default)
No environment variables needed. The application will use:
- Server Address: `localhost:7233`
- Namespace: `default`
- Task Queue: `hello-world-task-queue`

### Temporal Cloud
Required environment variables:
```bash
# Use either TEMPORAL_HOST_URL or TEMPORAL_HOST_ADDRESS
export TEMPORAL_HOST_URL=<namespace>.<account_id>.tmprl.cloud:7233
# or
export TEMPORAL_HOST_ADDRESS=<namespace>.<account_id>.tmprl.cloud:7233

export TEMPORAL_NAMESPACE=<namespace>.<account_id>
export TEMPORAL_TLS_CERT=<path-to-cert.pem>
export TEMPORAL_TLS_KEY=<path-to-cert.key>
export TEMPORAL_TASK_QUEUE=custom-task-queue  # Optional
```

## Running Locally

### 1. Start Temporal Server

Start the Temporal development server with Web UI:

```bash
temporal server start-dev --ui-port 8080
```

This starts:
- Temporal Server on port 7233 (default)
- Web Interface at http://localhost:8080

### 2. Run the Worker

In one terminal:
```bash
go run .
```

### 3. Execute the Workflow

In another terminal:
```bash
# Basic execution with default name "Temporal"
go run . start

# Or with a custom name
go run . start YourName
```

## Temporal Cloud Setup

### 1. Install Temporal CLI
Install the Temporal CLI by following the [official documentation](https://docs.temporal.io/cli).

### 2. Generate mTLS Certificate
Choose one of these approaches:
- Generate certificates through Temporal Cloud Console
- Use your own certificates (must meet Temporal's requirements)

### 3. Create Namespace
Create a namespace in Temporal Cloud Console (e.g., `default`, `dev-namespace`, `prod-namespace`)

### 4. Verify Connection
Test your Temporal Cloud connection:

```bash
temporal workflow list \
  --address <namespace>.<account_id>.tmprl.cloud:7233 \
  --namespace <namespace>.<account_id> \
  --tls-cert-path <cert.pem> \
  --tls-key-path <cert.key>
```

A new namespace should return an empty list.

### 5. Configure Environment
Set up your environment variables:

```bash
# Temporal Cloud Configuration
export TEMPORAL_HOST_URL=<namespace>.<account_id>.tmprl.cloud:7233  # Preferred
# or
export TEMPORAL_HOST_ADDRESS=<namespace>.<account_id>.tmprl.cloud:7233

export TEMPORAL_NAMESPACE=<namespace>.<account_id>
export TEMPORAL_TLS_CERT=<cert.pem>
export TEMPORAL_TLS_KEY=<cert.key>
```

## Stack

* Go 1.16+
* [Temporal SDK](https://github.com/temporalio/sdk-go) for workflow orchestration

## License

MIT 