Running two containerized instances for sharing tests

Overview

- This repository includes a `Dockerfile`, `docker-compose.yml`, and a helper script to run two isolated instances of the app for testing sharing/transfer.

Quick start

1. Build and run both instances:

```bash
./scripts/run-two-containers.sh
```

2. Tail logs (both instances):

```bash
docker compose logs -f
```

Storage

- Each instance uses a bind-mounted directory under `./tmp/instance-a` and `./tmp/instance-b` respectively. Secrets and local state are stored under those directories inside the container at `/app` and are isolated.

Discovery and networking

- The app uses a LAN transport which relies on advertisement and direct TCP/UDP dialing. In Docker the simplest approach is to use the docker bridge network created by `docker-compose`.
- Important options:
  - Bridge network (default): Containers can reach each other by container name on the user-defined bridge network. `docker-compose` exposes ports to the host; internal container ports are reachable by other containers using the internal network.
  - Published ports: We also publish ports 39001 and 39002 to the host for convenience; the app may use a discovery advertisement containing an address. If the app advertises the host port, receivers on other containers may not reach it unless the advertisement points to the container reachable address.
  - Host networking: Using `network_mode: host` removes NAT and makes discovery trivial because containers share the host network; this is not available on Docker Desktop for macOS and has security implications.

Recommended approach for discovery inside Docker Compose

- Use the internal service DNS name and listen on fixed ports inside container(s). In `docker-compose.yml` we set environment variables `VPM_LISTEN_ADDR` to distinct listen ports for each container; inside the app you should prefer the listen address from `VPM_LISTEN_ADDR` environment variable to advertise.
- If the app advertises a network address for discovery, ensure it advertises the reachable container address (e.g., the compose service name and port or the host IP and published port). The simplest change is to make the app advertise the container listen address (0.0.0.0:port) and, if needed, override the advertised address via an environment variable for container runs.

Next steps

- If you want, I can:
  - Add environment variable support to the app to override the advertised discovery address (recommended), or
  - Update `docker-compose.yml` to use `network_mode: host` when running on Linux for simplest discovery.
