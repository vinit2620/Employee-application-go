# Employee App (Go + PostgreSQL on Kubernetes)

A containerized REST API built in Go backed by PostgreSQL, deployed on a local Kubernetes cluster (`kind`) with persistent storage, health/readiness probes, and multi-replica scaling.

## Architecture

- **Application:** Go HTTP server (`employee-app`) running 2 replicas behind a ClusterIP Service (`port 80 -> 8080`).
- **Database:** PostgreSQL 16 (`postgres:16-alpine`) managed by a Deployment using the `Recreate` strategy and a `1Gi` PersistentVolumeClaim (`postgres-pvc`).
- **Configuration & Secrets:** Database credentials injected via Kubernetes Secrets (`postgres-secret`).
- **Probes:** `/healthz` (liveness) and `/readyz` (readiness, verifies DB connectivity) on the app; `pg_isready` readiness probe on PostgreSQL.

## Project Structure

```text
employee-app/
├── main.go                 # Go API & PostgreSQL connection logic
├── go.mod / go.sum         # Go module dependencies
├── Dockerfile              # Multi-stage Docker build
├── .gitignore              # Git ignore rules
├── README.md               # Project documentation
└── k8s/
    ├── app-deployment.yaml # employee-app Deployment & Service
    ├── db-deployment.yaml  # PostgreSQL Deployment & PersistentVolumeClaim
    ├── db-service.yaml     # PostgreSQL ClusterIP Service
    └── db-secret.yaml      # PostgreSQL credentials Secret
```

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Quick Start

### 1. Create a local `kind` cluster

```bash
kind create cluster --name kind
kubectl config use-context kind-kind
```

### 2. Build the Docker image and load it into `kind`

```bash
docker build -t employee-app:v1 .
kind load docker-image employee-app:v1 --name kind
```

### 3. Deploy the Kubernetes manifests

```bash
kubectl apply -f k8s/
```

Verify that the PVC is `Bound` and all pods reach `1/1 Running`:

```bash
kubectl get pvc
kubectl get pods -w
```

### 4. Port-forward the Service

```bash
kubectl port-forward svc/employee-app 8082:80
```

### 5. Test the API

In a second terminal:

```bash
# Health & readiness checks
curl http://localhost:8082/healthz
curl http://localhost:8082/readyz

# Create an employee
curl -X POST http://localhost:8082/employees \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","department":"SRE"}'

# List all employees
curl http://localhost:8082/employees
```

## Verifying Data Persistence

Confirm that PostgreSQL data survives a pod crash or deletion:

```bash
# Delete the running Postgres pod
kubectl delete pod -l app=postgres

# Wait for the replacement pod to become ready
kubectl wait --for=condition=ready pod -l app=postgres --timeout=60s

# Verify the employee record is still returned
curl http://localhost:8082/employees
```

## Troubleshooting (Linux / Ubuntu)

If `kube-proxy`, `coredns`, or `local-path-provisioner` fail with `too many open files` or `i/o timeout` on `10.96.0.1:443`, increase your host `inotify` limits:

```bash
sudo sysctl fs.inotify.max_user_watches=524288
sudo sysctl fs.inotify.max_user_instances=512
```

To make the change persist across reboots:

```bash
echo "fs.inotify.max_user_watches = 524288" | sudo tee -a /etc/sysctl.conf
echo "fs.inotify.max_user_instances = 512" | sudo tee -a /etc/sysctl.conf
```

## Cleanup

```bash
kubectl delete -f k8s/
kind delete cluster --name kind
```