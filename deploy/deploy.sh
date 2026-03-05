#!/bin/bash
# Fingerprint Gateway Deployment Script
# Supports Docker Compose and Kubernetes deployments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
VERSION=${VERSION:-v3.0.0}
REGISTRY=${REGISTRY:-docker.io}
IMAGE_NAME=${IMAGE_NAME:-vistone/fingerprint}
NAMESPACE=${NAMESPACE:-fingerprint}

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Resolve docker compose command and run it consistently.
compose_cmd() {
    if command -v docker-compose > /dev/null 2>&1; then
        echo "docker-compose"
        return 0
    fi

    if command -v docker > /dev/null 2>&1 && docker compose version > /dev/null 2>&1; then
        echo "docker compose"
        return 0
    fi

    return 1
}

run_compose() {
    local cmd
    if ! cmd=$(compose_cmd); then
        log_error "Docker Compose not found. Install docker-compose or Docker Compose plugin."
        return 1
    fi

    if [[ "$cmd" == "docker compose" ]]; then
        docker compose "$@"
    else
        docker-compose "$@"
    fi
}

check_compose_access() {
    local compose_file="$1"

    if [[ ! -r "$compose_file" ]]; then
        log_error "Cannot read compose file: $compose_file"
        return 1
    fi

    local out
    if ! out=$(run_compose -f "$compose_file" config 2>&1); then
        if echo "$out" | grep -qi "permission denied"; then
            log_error "Docker Compose cannot access $compose_file"
            log_warn "Detected permission issue (often caused by snap docker with /media paths)."
            log_info "Try: sudo snap connect docker:removable-media"
            log_info "Or move project to a directory under your home (e.g. ~/projects)."
        else
            log_error "Docker Compose precheck failed:"
            echo "$out"
        fi
        return 1
    fi
}

# Show usage
usage() {
    cat << EOF
Fingerprint Gateway Deployment Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    build           Build Docker image
    push            Push Docker image to registry
    deploy-docker   Deploy using Docker Compose
    deploy-k8s      Deploy to Kubernetes
    undeploy        Remove deployment
    status          Check deployment status
    logs            Show logs
    scale           Scale deployment (k8s only)
    test            Run smoke tests
    help            Show this help message

Options:
    -v, --version   Set version tag (default: $VERSION)
    -r, --registry  Set image registry (default: $REGISTRY)
    -n, --namespace Set Kubernetes namespace (default: $NAMESPACE)

Examples:
    $0 build --version v3.0.0
    $0 deploy-docker
    $0 deploy-k8s --namespace production
    $0 scale --replicas 10
EOF
}

# Build Docker image
build_image() {
    log_info "Building Docker image $REGISTRY/$IMAGE_NAME:$VERSION..."

    local script_dir repo_root dockerfile_path
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    repo_root="$(cd "$script_dir/.." && pwd)"
    dockerfile_path="$script_dir/docker/Dockerfile"

    # Check if Dockerfile exists
    if [[ ! -f "$dockerfile_path" ]]; then
        log_error "Dockerfile not found in docker/"
        exit 1
    fi

    # Prefer multi-platform build when the current buildx driver supports it.
    local buildx_driver=""
    if command -v docker > /dev/null 2>&1 && docker buildx version > /dev/null 2>&1; then
        buildx_driver=$(docker buildx inspect 2>/dev/null | awk -F': ' '/^Driver:/ {print $2; exit}')
    fi

    if [[ -n "$buildx_driver" && "$buildx_driver" != "docker" ]]; then
        docker buildx build \
            --platform linux/amd64,linux/arm64 \
            -t $REGISTRY/$IMAGE_NAME:$VERSION \
            -t $REGISTRY/$IMAGE_NAME:latest \
            -f "$dockerfile_path" \
            --push=false \
            "$repo_root"
    else
        if [[ "$buildx_driver" == "docker" ]]; then
            log_warn "buildx driver 'docker' does not support local multi-platform builds."
        else
            log_warn "buildx is unavailable; falling back to single-platform build."
        fi

        docker build \
            -t $REGISTRY/$IMAGE_NAME:$VERSION \
            -t $REGISTRY/$IMAGE_NAME:latest \
            -f "$dockerfile_path" \
            "$repo_root"
    fi

    log_info "Image built successfully"
}

# Push image to registry
push_image() {
    log_info "Pushing image to $REGISTRY..."

    docker push $REGISTRY/$IMAGE_NAME:$VERSION
    docker push $REGISTRY/$IMAGE_NAME:latest

    log_info "Image pushed successfully"
}

# Deploy with Docker Compose
deploy_docker() {
    log_info "Deploying with Docker Compose..."

    cd docker

    # Update image tag
    sed -i.bak "s|image: vistone/fingerprint:.*|image: $REGISTRY/$IMAGE_NAME:$VERSION|" docker-compose.yml
    rm -f docker-compose.yml.bak

    # Verify compose can read files before starting services.
    check_compose_access "docker-compose.yml"

    # Start services
    run_compose up -d

    # Wait for health check
    log_info "Waiting for services to be ready..."
    sleep 5

    # Health check
    if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
        log_info "Services are ready!"
        log_info "API available at: http://localhost:8080"
    else
        log_warn "Services may not be ready yet. Check logs with: $0 logs"
    fi

    cd ..
}

# Deploy to Kubernetes
deploy_k8s() {
    log_info "Deploying to Kubernetes namespace $NAMESPACE..."

    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not found. Please install kubectl."
        exit 1
    fi

    # Check cluster connection
    if ! kubectl cluster-info > /dev/null 2>&1; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    cd kubernetes

    # Update image in kustomization
    kustomize edit set image vistone/fingerprint=$REGISTRY/$IMAGE_NAME:$VERSION

    # Apply manifests
    kustomize build . | kubectl apply -f -

    # Wait for deployment
    log_info "Waiting for deployment to be ready..."
    kubectl -n $NAMESPACE rollout status deployment/fingerprint-gateway --timeout=5m

    log_info "Deployment complete!"
    log_info "Run '$0 status' to check deployment status"

    cd ..
}

# Undeploy
undeploy() {
    log_info "Removing deployment..."

    if [[ -f "kubernetes/kustomization.yaml" ]]; then
        cd kubernetes
        kustomize build . | kubectl delete -f - --ignore-not-found
        cd ..
    fi

    if [[ -f "docker/docker-compose.yml" ]]; then
        cd docker
        run_compose down -v
        cd ..
    fi

    log_info "Deployment removed"
}

# Check status
status() {
    log_info "Checking deployment status..."

    if kubectl get ns $NAMESPACE > /dev/null 2>&1; then
        echo ""
        echo "=== Kubernetes Status ==="
        kubectl -n $NAMESPACE get pods,svc,hpa
        echo ""
        echo "=== Pod Details ==="
        kubectl -n $NAMESPACE get pods -o wide
    fi

    echo ""
    echo "=== Docker Compose Status ==="
    cd docker && run_compose ps 2>/dev/null || echo "Not running"
    cd ..
}

# Show logs
logs() {
    local follow=""

    if [[ "$1" == "-f" ]]; then
        follow="-f"
    fi

    if kubectl get ns $NAMESPACE > /dev/null 2>&1; then
        kubectl -n $NAMESPACE logs -l app=fingerprint-gateway $follow
    else
        cd docker && run_compose logs $follow
        cd ..
    fi
}

# Scale deployment
scale() {
    local replicas=${1:-5}

    log_info "Scaling deployment to $replicas replicas..."
    kubectl -n $NAMESPACE scale deployment fingerprint-gateway --replicas=$replicas
    kubectl -n $NAMESPACE rollout status deployment/fingerprint-gateway
}

# Run smoke tests
test_deployment() {
    log_info "Running smoke tests..."

    local base_url="${FP_URL:-http://localhost:8080}"
    local failed=0

    # Test health endpoint
    echo -n "Testing health endpoint... "
    if curl -sf "$base_url/health" > /dev/null; then
        echo "OK"
    else
        echo "FAILED"
        failed=$((failed + 1))
    fi

    # Test API endpoint
    echo -n "Testing classify endpoint... "
    if curl -sf -X POST "$base_url/api/v1/classify" \
        -H "Content-Type: application/json" \
        -d '{"tls_version":"1.3","cipher_suites":[49195]}' > /dev/null; then
        echo "OK"
    else
        echo "FAILED"
        failed=$((failed + 1))
    fi

    # Test metrics endpoint
    echo -n "Testing metrics endpoint... "
    if curl -sf "$base_url/metrics" > /dev/null; then
        echo "OK"
    else
        echo "FAILED"
        failed=$((failed + 1))
    fi

    echo ""
    if [[ $failed -eq 0 ]]; then
        log_info "All tests passed!"
        return 0
    else
        log_error "$failed test(s) failed"
        return 1
    fi
}

# Main function
main() {
    local command=${1:-help}
    if [[ $# -gt 0 ]]; then
        shift
    fi

    # Parse options
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -r|--registry)
                REGISTRY="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            --replicas)
                REPLICAS="$2"
                shift 2
                ;;
            -f)
                FOLLOW="-f"
                shift
                ;;
            *)
                shift
                ;;
        esac
    done

    case $command in
        build)
            build_image
            ;;
        push)
            push_image
            ;;
        deploy-docker)
            deploy_docker
            ;;
        deploy-k8s)
            deploy_k8s
            ;;
        undeploy)
            undeploy
            ;;
        status)
            status
            ;;
        logs)
            logs $FOLLOW
            ;;
        scale)
            scale ${REPLICAS:-5}
            ;;
        test)
            test_deployment
            ;;
        help|--help|-h)
            usage
            ;;
        *)
            log_error "Unknown command: $command"
            usage
            exit 1
            ;;
    esac
}

main "$@"
