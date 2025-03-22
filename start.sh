#!/bin/bash

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Store the process group ID for later cleanup
SCRIPT_PGID=$$

# Function to cleanup processes
cleanup() {
    echo -e "\n${BLUE}Initiating graceful shutdown...${NC}"
    
    if [ ! -z "$WORKER_PID" ]; then
        echo -e "${YELLOW}Stopping worker process with PID $WORKER_PID...${NC}"
        kill -SIGTERM $WORKER_PID 2>/dev/null || true
        
        # Wait for graceful shutdown
        for i in {1..3}; do
            if ! kill -0 $WORKER_PID 2>/dev/null; then
                echo -e "${GREEN}Worker stopped gracefully${NC}"
                break
            fi
            sleep 1
        done
        
        # Force kill if still running
        if kill -0 $WORKER_PID 2>/dev/null; then
            echo -e "${YELLOW}Force stopping worker...${NC}"
            kill -9 $WORKER_PID 2>/dev/null || true
        fi
    fi

    # Very aggressive cleanup to kill ALL related processes
    echo -e "${YELLOW}Performing complete cleanup of all related processes...${NC}"
    
    # Kill by various patterns
    pkill -f "go run ./cmd/helloworld" 2>/dev/null || true
    pkill -f "cmd/helloworld" 2>/dev/null || true
    pkill -f "__debug_bin" 2>/dev/null || true
    pkill -f "go-build" 2>/dev/null || true
    
    # Short pause
    sleep 1
    
    # Force kill any remaining processes
    pkill -9 -f "go run ./cmd/helloworld" 2>/dev/null || true
    pkill -9 -f "cmd/helloworld" 2>/dev/null || true
    pkill -9 -f "__debug_bin" 2>/dev/null || true
    
    # Find and kill all processes that might be related to our app
    for pid in $(ps -ef | grep -E '[g]o .*/cmd/helloworld|[c]md/helloworld|[h]ello-world' | awk '{print $2}'); do
        echo -e "${RED}Killing process $pid${NC}"
        kill -9 $pid 2>/dev/null || true
    done
    
    # Give processes time to terminate
    sleep 2
    
    # Final check to ensure no processes remain
    REMAINING=$(ps -ef | grep -E '[g]o .*/cmd/helloworld|[c]md/helloworld|[h]ello-world' | wc -l)
    if [ $REMAINING -gt 0 ]; then
        echo -e "${RED}WARNING: $REMAINING processes still running. Attempting harder kill...${NC}"
        # Try killall as a last resort
        killall -9 go 2>/dev/null || true
        # Kill any other processes this script spawned
        pkill -9 -P $SCRIPT_PGID 2>/dev/null || true
    fi
    
    echo -e "${GREEN}Cleanup completed${NC}"
}

# Function to check worker startup
check_worker() {
    local pid=$1
    local timeout=5
    local count=0
    
    echo -e "${YELLOW}Waiting for worker initialization...${NC}"
    while [ $count -lt $timeout ]; do
        if ! kill -0 $pid 2>/dev/null; then
            echo -e "\n${RED}✗ Worker failed to start${NC}"
            return 1
        fi
        sleep 1
        echo -n "."
        count=$((count + 1))
    done
    echo -e "\n${GREEN}✓ Worker initialized${NC}"
    return 0
}

# Set up trap for Ctrl+C (SIGINT) and SIGTERM, but not EXIT
# We want to control when cleanup happens
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
        echo -e "${YELLOW}Attempt $((count + 1))/$retries: Waiting for $host:$port...${NC}"
        sleep $wait_time
        count=$((count + 1))
    done
    return 1
}

# Function to check if we're using Temporal Cloud
is_temporal_cloud() {
    [ ! -z "$TEMPORAL_HOST_URL" ] || [ ! -z "$TEMPORAL_HOST_ADDRESS" ]
}

# Print header
echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║${GREEN}    Temporal Hello World Demo Application    ${BLUE}║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# OpenTelemetry Configuration
export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4317"
export OTEL_RESOURCE_ATTRIBUTES="service.name=temporal-hello-world"

# If SigNoz ingestion key is provided, set it
if [ ! -z "$SIGNOZ_INGESTION_KEY" ]; then
    export OTEL_EXPORTER_OTLP_HEADERS="signoz-ingestion-key=$SIGNOZ_INGESTION_KEY"
    echo -e "${GREEN}✓ Using SigNoz with provided ingestion key${NC}"
fi

# Check Temporal Server (only for local development)
if ! is_temporal_cloud; then
    echo -e "\n${BLUE}Checking Temporal Server...${NC}"
    if ! check_port localhost 7233 3 2; then
        echo -e "${RED}✗ Error: Local Temporal Server is not running${NC}"
        echo -e "Please start it first with:"
        echo -e "${GREEN}temporal server start-dev --ui-port 8080${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Local Temporal Server is running${NC}"
else
    echo -e "\n${BLUE}Checking Temporal Cloud configuration...${NC}"
    if [ -z "$TEMPORAL_TLS_CERT" ] || [ -z "$TEMPORAL_TLS_KEY" ]; then
        echo -e "${RED}✗ Error: TEMPORAL_TLS_CERT and TEMPORAL_TLS_KEY must be set for Temporal Cloud${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Temporal Cloud credentials verified${NC}"
fi

# Check SigNoz/OpenTelemetry Collector
echo -e "\n${BLUE}Checking OpenTelemetry Collector...${NC}"
if ! check_port localhost 4317 3 2; then
    echo -e "${YELLOW}⚠ Warning: OpenTelemetry Collector is not running${NC}"
    echo -e "   Metrics and traces will not be exported"
else
    echo -e "${GREEN}✓ OpenTelemetry Collector is running${NC}"
fi

# Clean existing processes before starting
echo -e "\n${BLUE}Preparing environment...${NC}"
cleanup
sleep 2

# Start the worker
echo -e "\n${BLUE}Starting worker process...${NC}"
go run ./cmd/helloworld -worker &
WORKER_PID=$!

# Check if worker started successfully
if ! check_worker $WORKER_PID; then
    echo -e "${RED}Error: Worker failed to start. Check the logs above for details.${NC}"
    exit 1
fi

echo -e "\n${GREEN}✓ Worker is running${NC}"

# Start the workflow
echo -e "\n${BLUE}Executing workflow...${NC}"
if WORKFLOW_NAME="Temporal" go run ./cmd/helloworld; then
    echo -e "${GREEN}✓ Workflow completed successfully${NC}"
else
    echo -e "${RED}✗ Workflow execution failed${NC}"
    exit 1
fi

# Let the worker run for a while to send metrics
echo -e "\n${BLUE}Keeping worker alive to send metrics to SigNoz...${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop the demo at any time${NC}"

# Progress bar for waiting, longer duration
METRICS_WAIT_TIME=60  # Keep worker running for 60 seconds to generate metrics
echo -e "Keeping worker alive for ${METRICS_WAIT_TIME} seconds to generate metrics..."

# Show a progress bar
for i in $(seq 1 $METRICS_WAIT_TIME); do
    # Update progress every 5 seconds
    if [ $((i % 5)) -eq 0 ]; then
        PERCENT=$((i * 100 / METRICS_WAIT_TIME))
        echo -ne "\r[${YELLOW}$PERCENT%${NC}] Running worker: $i/$METRICS_WAIT_TIME seconds"
    fi
    sleep 1
done
echo -e "\n${GREEN}✓ Metrics collection period completed${NC}"

# Now run cleanup to shut everything down
echo -e "\n${BLUE}Demo completed. Shutting down...${NC}"
cleanup

# Final sanity check before exiting
REMAINING=$(ps -ef | grep -E '[g]o .*/cmd/helloworld|[c]md/helloworld|[h]ello-world' | wc -l)
if [ $REMAINING -gt 0 ]; then
    echo -e "${RED}WARNING: Still found $REMAINING processes running. Killing all Go processes...${NC}"
    ps -ef | grep -E '[g]o .*/cmd/helloworld|[c]md/helloworld|[h]ello-world'
    killall -9 go 2>/dev/null || true
    pkill -9 -f "helloworld" 2>/dev/null || true
fi

echo -e "\n${GREEN}✓ Demo completed successfully${NC}"
echo -e "${BLUE}════════════════════════════════════════════${NC}"
exit 0 