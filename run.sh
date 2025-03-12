#!/bin/bash

# Function to cleanup processes
cleanup() {
    echo -e "\nCleaning up processes..."
    pkill -f "go run ." 2>/dev/null || true
    exit 0
}

# Set up trap for Ctrl+C (SIGINT) and SIGTERM
trap cleanup SIGINT SIGTERM

# Function to check if a port is open
check_port() {
    local host=$1
    local port=$2
    local retries=$3
    local wait_time=$4
    local count=0

    while [ $count -lt $retries ]; do
        nc -z $host $port > /dev/null 2>&1
        if [ $? -eq 0 ]; then
            return 0
        fi
        echo "Attempt $((count + 1))/$retries: Waiting for $host:$port..."
        sleep $wait_time
        count=$((count + 1))
    done
    return 1
}

# Function to check if we're using Temporal Cloud
is_temporal_cloud() {
    [ ! -z "$TEMPORAL_HOST_URL" ] || [ ! -z "$TEMPORAL_HOST_ADDRESS" ]
}

echo "Starting application..."

# Check Temporal Server (only for local development)
if ! is_temporal_cloud; then
    echo "Checking local Temporal Server..."
    if ! check_port localhost 7233 3 2; then
        echo "ERROR: Local Temporal Server is not running. Please start it first:"
        echo "temporal server start-dev --ui-port 8080"
        exit 1
    fi
else
    echo "Using Temporal Cloud configuration..."
    # Verify required environment variables
    if [ -z "$TEMPORAL_TLS_CERT" ] || [ -z "$TEMPORAL_TLS_KEY" ]; then
        echo "ERROR: TEMPORAL_TLS_CERT and TEMPORAL_TLS_KEY must be set for Temporal Cloud"
        exit 1
    fi
fi

# Clean existing processes
echo "Cleaning up existing processes..."
pkill -f "go run ." 2>/dev/null || true
sleep 2

# Start the worker
echo "Starting worker..."
go run . &
WORKER_PID=$!

# Wait for worker to initialize
echo "Waiting for worker to initialize..."
sleep 5

# Start the workflow
echo "Starting workflow with default name 'Temporal'..."
go run . start "Temporal"

echo "Workflow completed."

# Cleanup at the end
cleanup 