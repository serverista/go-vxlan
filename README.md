# Go VXLAN Implementation

This repository demonstrates how to programmatically create and configure a point-to-point VXLAN (Virtual eXtensible Local Area Network) overlay network using Golang. It includes a containerized working example managed by Docker Compose for end-to-end testing.

## Prerequisites

- Docker
- Docker Compose v2 (installed via `docker compose` or `docker-compose`)

## Core Components

- **`main.go`**: The core Golang application. It uses the `github.com/vishvananda/netlink` library to interact with the Linux kernel network stack. It resolves the underlying network interface, establishes a new `netlink.Vxlan` tunnel to a designated remote IP address, assigns it an overlay IP address, and brings the link up.
- **`Dockerfile`**: A multi-stage build that compiles the Golang code and packages it inside an `alpine` layer. It additionally installs `iproute2` and `iputils` for standard network debugging (`ip addr`, `ping`). 
- **`docker-compose.yml`**: Provisions a mock 2-node physical network bridging structure (the `underlay`). It sets up symmetrical environment variables so that `node1` (`172.20.0.2`) targets `node2` (`172.20.0.3`) as its destination VTEP (Virtual Tunnel Endpoint), forming an isolated tunnel subnet over `10.0.0.0/24`.
- **`test.sh`**: A comprehensive bash script to automatically build, orchestrate, test, and tear down the demonstration.

## How It Works

The Go program maps the VXLAN configuration entirely through standard environment variables injected at runtime:
- `VXLAN_ID`: The Virtual Network Identifier (VNI) isolating the traffic (e.g., `100`).
- `LOCAL_IP`: The IP address of the local machine/container acting as the physical bridging interface (underlay).
- `REMOTE_IP`: The corresponding IP address of the remote peer/container completing the circuit.
- `VXLAN_IP`: The logical IP address (with CIDR suffix) assigned to the newly established VXLAN interface passing traffic through the tunnel (e.g., `10.0.0.1/24`).

*Note*: Because this application directly communicates with Linux via netlink to create and assign physical network boundaries, the Docker containers run explicitly with the `NET_ADMIN` capability.

## Running the Example

Simply execute the included test script from the project root to observe the orchestration:

```bash
chmod +x test.sh
./test.sh
```

### What to expect

1. The orchestrator will build the container image hosting the Go binary.
2. Both Docker nodes will align to the custom `underlay` network.
3. The executing Go routines will mutually bind the `vxlan100` interfaces targeting each other.
4. The system validates the tunnel by pinging the payload VXLAN IP `10.0.0.2` (on `node2`) originating directly from `node1`.
5. Upon confirmation of ICMP packet routing across the overlay boundary, `docker compose` terminates and cleans up the mock deployment stack.
