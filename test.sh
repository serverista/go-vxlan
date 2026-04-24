#!/bin/bash
set -e

echo "Building and starting Docker containers..."
docker compose build
docker compose up -d

echo "Waiting for containers and VXLAN interfaces to initialize..."
sleep 5

echo "Dumping network interfaces on node1 for debugging..."
docker exec go-vxlan-node1-1 ip addr show

echo "Testing connectivity from node1 to node2 over VXLAN (10.0.0.1 -> 10.0.0.2)..."
docker exec go-vxlan-node1-1 ping -c 3 10.0.0.2

echo "Test successful."

echo "Tearing down..."
docker compose down
