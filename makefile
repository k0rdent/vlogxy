LOCALBIN ?= $(shell pwd)/bin
export LOCALBIN
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

TEMPLATES_DIR := helm-charts

DEPLOY_NAMESPACE ?= kof
CONTAINER_TOOL ?= docker
KIND_NETWORK ?= kind

KIND_CLUSTER_NAME ?= kcm-dev

dev:
	mkdir -p dev
lint-chart-%:
	$(HELM) dependency update $(TEMPLATES_DIR)/$*
	$(HELM) lint --strict $(TEMPLATES_DIR)/$* --set global.lint=true

.PHONY: dev-deploy
dev-deploy: dev dev-build ## Deploy vlogxy helm chart to the K8s cluster specified in ~/.kube/config
	cp -f $(TEMPLATES_DIR)/vlogxy/values.yaml dev/vlogxy-values.yaml
	@$(YQ) eval -i '.image.registry = "docker.io/library"' dev/vlogxy-values.yaml # See `load docker-image`
	@$(YQ) eval -i '.image.repository = "vlogxy"' dev/vlogxy-values.yaml
	$(HELM_UPGRADE) --create-namespace -n $(DEPLOY_NAMESPACE) vlogxy ./$(TEMPLATES_DIR)/vlogxy -f dev/vlogxy-values.yaml

.PHONY: dev-build
dev-build: docker-build ## Build vlogxy docker image
	@vlogxy_version=v$$($(YQ) .version $(TEMPLATES_DIR)/vlogxy/Chart.yaml); \
	$(CONTAINER_TOOL) tag vlogxy vlogxy:$$vlogxy_version; \
	$(KIND) load docker-image vlogxy:$$vlogxy_version --name $(KIND_CLUSTER_NAME)

.PHONY: docker-build
docker-build: dev yq goreleaser ## Build docker image
	@ \
	cp -f .goreleaser.yml dev/.goreleaser.yml; \
	ARCH=$$(uname -m); \
	if [ "$$ARCH" = "arm64" ]; then \
		$(YQ) eval -i 'del(.builds[0])' dev/.goreleaser.yml; \
		$(YQ) eval -i 'del(.dockers[0])' dev/.goreleaser.yml; \
		$(YQ) eval -i '.builds[0].dir = "."' dev/.goreleaser.yml; \
		$(YQ) eval -i '.dockers[0].skip_push = "true"' dev/.goreleaser.yml; \
		$(YQ) eval -i '.dockers[0].dockerfile = "./goreleaser.dockerfile"' dev/.goreleaser.yml; \
		$(YQ) eval -i '.dockers[0].image_templates[1] = "vlogxy:latest"' dev/.goreleaser.yml; \
	elif [ "$$ARCH" = "x86_64" ]; then \
		$(YQ) eval -i 'del(.builds[1])' dev/.goreleaser.yml; \
		$(YQ) eval -i 'del(.dockers[1])' dev/.goreleaser.yml; \
		$(YQ) eval -i '.builds[0].dir = "."' dev/.goreleaser.yml; \
		$(YQ) eval -i '.dockers[0].skip_push = "true"' dev/.goreleaser.yml; \
		$(YQ) eval -i '.dockers[0].dockerfile = "./goreleaser.dockerfile"' dev/.goreleaser.yml; \
		$(YQ) eval -i '.dockers[0].image_templates[1] = "vlogxy:latest"' dev/.goreleaser.yml; \
	fi; \
	IMAGE_REPO=vlogxy GITHUB_OWNER=k0rdent GITHUB_REPO_NAME=vlogxy VERSION=latest $(GORELEASER) release --snapshot --clean -f dev/.goreleaser.yml


## Tool Versions
HELM_VERSION ?= v3.18.5
YQ_VERSION ?= v4.44.2
KIND_VERSION ?= v0.27.0
GORELEASER_VERSION ?= v2.10.2
GOLANGCI_LINT_VERSION ?= v2.5.0

## Tool Binaries
HELM ?= $(LOCALBIN)/helm-$(HELM_VERSION)
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
GORELEASER ?= $(LOCALBIN)/goreleaser
KIND ?= $(LOCALBIN)/kind-$(KIND_VERSION)
YQ ?= $(LOCALBIN)/yq-$(YQ_VERSION)

HELM_UPGRADE = $(HELM) upgrade -i --reset-values --wait
KUBECTL ?= kubectl
export HELM HELM_UPGRADE
export YQ

.PHONY: yq
yq: $(YQ) ## Download yq locally if necessary.
$(YQ): | $(LOCALBIN)
	$(call go-install-tool,$(YQ),github.com/mikefarah/yq/v4,${YQ_VERSION})

.PHONY: kind
kind: $(KIND) ## Download kind locally if necessary.
$(KIND): | $(LOCALBIN)
	$(call go-install-tool,$(KIND),sigs.k8s.io/kind,${KIND_VERSION})

.PHONY: helm
helm: $(HELM) ## Download helm locally if necessary.
HELM_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3"
$(HELM): | $(LOCALBIN)
	rm -f $(LOCALBIN)/helm-*
	curl -s --fail $(HELM_INSTALL_SCRIPT) | USE_SUDO=false HELM_INSTALL_DIR=$(LOCALBIN) DESIRED_VERSION=$(HELM_VERSION) BINARY_NAME=helm-$(HELM_VERSION) PATH="$(LOCALBIN):$(PATH)" bash

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: goreleaser
goreleaser: $(GORELEASER) ## Download goreleaser locally if necessary
$(GORELEASER): $(LOCALBIN)
	$(call go-install-tool,$(GORELEASER),github.com/goreleaser/goreleaser/v2,$(GORELEASER_VERSION))

.PHONY: cli-install
cli-install: yq helm kind ## Install the necessary CLI tools for deployment, development and testing.

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

.PHONY: test
test: lint
	go test -v $$(go list ./... | grep -v /e2e)


# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
if [ ! -f $(1) ]; then mv -f "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1); fi ;\
}
endef


