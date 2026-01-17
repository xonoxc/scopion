#!/bin/bash

echo "Checking for multiple service processes..."
echo "=========================================="
echo "=========================================="
echo "=========================================="
echo "=========================================="

# Services being monitored by scopion
services=("api" "worker" "webhook" "cron" "scheduler" "auth" "payment")

for service in "${services[@]}"; do
    # Count processes (excluding grep itself)
    count=$(ps aux | grep -E "(^|/)$service([^/]|$)" | grep -v grep | wc -l)

    if [ "$count" -eq 0 ]; then
        echo "$service: No processes running"
    elif [ "$count" -eq 1 ]; then
        echo "$service: 1 process running"
    else
        echo "$service: $count processes running (multiple instances!)"
    fi
done

echo ""
echo "Scopion Dashboard:"
scopion_count=$(ps aux | grep './bin/scopion' | grep -v grep | wc -l)
if [ "$scopion_count" -eq 0 ]; then
    echo "❌ No scopion dashboard running"
elif [ "$scopion_count" -eq 1 ]; then
    echo "1 scopion dashboard running"
else
    echo "$scopion_count scopion dashboards running (multiple instances!)"
fi
