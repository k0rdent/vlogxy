# Variables
IMAGE_NAME ?= victorialogs-aggregator
IMAGE_TAG ?= latest
FULL_IMAGE_NAME = $(IMAGE_NAME):$(IMAGE_TAG)
KIND_CLUSTER_NAME ?= kcm-dev

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Main binary
BINARY_NAME=victorialogs-aggregator
MAIN_PATH=./cmd/vl-aggregator

.PHONY: all build clean test run docker-build docker-push kind-load helm-install help

all: test build

## build: Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH)

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## run: Run the application locally
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image: $(FULL_IMAGE_NAME)"
	docker build -t $(FULL_IMAGE_NAME) -f deploy/Dockerfile .

## docker-push: Push Docker image to registry
docker-push: docker-build
	@echo "Pushing Docker image: $(FULL_IMAGE_NAME)"
	docker push $(FULL_IMAGE_NAME)

## kind-load: Build and load Docker image into kind cluster
kind-load: docker-build
	@echo "Loading image $(FULL_IMAGE_NAME) into kind cluster: $(KIND_CLUSTER_NAME)"
	kind load docker-image $(FULL_IMAGE_NAME) --name $(KIND_CLUSTER_NAME)

## helm-install: Install Helm chart into kind cluster
helm-install:
	@echo "Installing Helm chart..."
	helm upgrade --install victorialogs-aggregator ./deploy/helm-chart -n kof \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG) \
		--set image.pullPolicy=IfNotPresent

## helm-uninstall: Uninstall Helm chart
helm-uninstall:
	@echo "Uninstalling Helm chart..."
	helm uninstall victorialogs-aggregator

## kind-deploy: Full deployment to kind (create cluster, build image, load to kind, install helm)
kind-deploy: kind-create kind-load helm-install
	@echo "Deployment to kind completed!"
	@echo "Check pods: kubectl get pods"
	@echo "Check service: kubectl get svc"

## kind-redeploy: Redeploy to existing kind cluster (build, load, upgrade helm)
kind-redeploy: kind-load helm-install
	@echo "Redeployment completed!"

## deploy: Build image, load to kind and upgrade helm (one command for quick iteration)
deploy: kind-load helm-install
	@echo "✅ Deploy completed!"
	@echo "Check status: kubectl get pods -w"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
