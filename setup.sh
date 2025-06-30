#!/bin/bash

# --- Distributed Key-Value Store Setup Script ---
# This script automates the setup process for the distributed key-value store
# as described in your README. It builds the Go application, starts two
# kvstore instances in the background, and then sends the join command.

# Function to clean up background processes on exit
cleanup() {
    echo -e "\n--- Cleaning up background kvstore processes ---"
    # Find processes started by this script and kill them
    # This assumes 'kvstore' is unique enough to identify the processes
    pkill -f "kvstore --node-id node1 --http-port 2222"
    pkill -f "kvstore --node-id node2 --http-port 2223"
    rm -rf /data # cleanup any data 
    rm kvstore
    echo "Cleanup complete. You can now close this terminal."
}

# Register the cleanup function to be called on script exit (Ctrl+C, etc.)
trap cleanup EXIT

echo "--- Starting Key-Value Store Setup ---"

rm -rf data

# 1. Build the Go application
echo "1. Building the kvstore application..."
go build -o kvstore .
if [ $? -ne 0 ]; then
    echo "Error: Go build failed. Please ensure Go is installed and your project is set up correctly."
    exit 1
fi
echo "kvstore built successfully."

# 2. Start kvstore instance 1 in the background
echo "2. Starting kvstore node1 (http:2222, raft:8222) in the background..."
./kvstore --node-id node1 --http-port 2222 --raft-port 8222 > /dev/null 2>&1 &
NODE1_PID=$!
echo "Node 1 started with PID: $NODE1_PID"

# 3. Start kvstore instance 2 in the background
echo "3. Starting kvstore node2 (http:2223, raft:8223) in the background..."
./kvstore --node-id node2 --http-port 2223 --raft-port 8223 > /dev/null 2>&1 &
NODE2_PID=$!
echo "Node 2 started with PID: $NODE2_PID"

# Give the nodes a moment to start up
echo "4. Giving nodes a few seconds to initialize..."
sleep 5

MAX_RETRIES=10
RETRY_COUNT=0
JOIN_SUCCESS=false

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do 
    # Capture both HTTP status and response body
    HTTP_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" "http://localhost:2222/Join?followerId=node2&followerAddr=localhost:8223")
    HTTP_BODY=$(echo "$HTTP_RESPONSE" | sed -e 's/HTTPSTATUS:.*//g')
    HTTP_STATUS=$(echo "$HTTP_RESPONSE" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

    echo "Join response (attempt $((RETRY_COUNT+1))): $HTTP_BODY (HTTP $HTTP_STATUS)"

    if [ "$HTTP_STATUS" -eq 200 ]; then
        echo -e "\nJoin request sent successfully."
        JOIN_SUCCESS=true
        break
    else
        RETRY_COUNT=$((RETRY_COUNT + 1))
        echo "Join request failed (HTTP $HTTP_STATUS). Retrying in 1 second..."
        sleep 1
    fi
done 

if [ "$JOIN_SUCCESS" = false ]; then
    echo "Error: Join request failed after $MAX_RETRIES attempts. Nodes might not be running or accessible."
    cleanup
    exit 1
fi

# # 5. Send the Join request from node2 to node1
# echo "5. Sending Join request for node2 to node1..."
# curl "http://localhost:2222/Join?followerId=node2&followerAddr=localhost:8223"
# if [ $? -ne 0 ]; then
#     echo "Error: Join request failed. Nodes might not be running or accessible."
#     # Attempt to kill processes even if join failed
#     cleanup
#     exit 1
# fi
echo -e "\nJoin request sent successfully. Setup is complete!"

echo -e "\n--- Interaction Instructions ---"
echo "You can now interact with your distributed key-value store."
echo "Open a NEW terminal and use the following commands:"
echo "To set a key-value pair (e.g., hello=world):"
echo "  curl \"http://localhost:2222/Set/hello/world\""
echo "To get the value for a key (e.g., hello):"
echo "  curl \"http://localhost:2222/Get/hello\""
echo ""
echo "You can use any of the node's HTTP ports (2222 or 2223) for Get requests."
echo "Set requests should generally go to the leader (initially node1 on port 2222)."
echo ""
echo "This terminal will remain open to keep the kvstore processes running."
echo "To stop the kvstore processes, simply press Ctrl+C in this terminal."

# Keep the script running until Ctrl+C is pressed
wait $NODE1_PID
wait $NODE2_PID
