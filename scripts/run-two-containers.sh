#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$ROOT_DIR"

if ! command -v docker >/dev/null 2>&1; then
	echo "docker not found in PATH. Install Docker Desktop or Docker Engine and retry."
	exit 1
fi

SERVICE_A=${1:-instance-a}
SERVICE_B=${2:-instance-b}

echo "Building docker image and starting two instances (services: $SERVICE_A, $SERVICE_B)..."
docker compose build --progress=plain
docker compose up -d --build

# wait for both containers to be running (timeout 30s)
echo "Waiting for containers to be healthy/running..."
for i in {1..30}; do
	a_status=$(docker inspect -f '{{.State.Running}}' "vpm_${SERVICE_A}" 2>/dev/null || echo false)
	b_status=$(docker inspect -f '{{.State.Running}}' "vpm_${SERVICE_B}" 2>/dev/null || echo false)
	if [ "$a_status" = "true" ] && [ "$b_status" = "true" ]; then
		echo "Both containers running" && break
	fi
	sleep 1
done

echo "To tail logs, run: docker compose logs -f $SERVICE_A $SERVICE_B"
docker compose logs -f $SERVICE_A $SERVICE_B

