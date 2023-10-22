IMG ?= dreamstax/kai-controller:$(VERSION)
VERSION ?= 0.0.1
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
ENVTEST_K8S_VERSION = 1.26.0
NAMESPACE ?= default
KIND_CLUSTER_NAME ?= kai
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
KIND ?=$(LOCALBIN)/kind
KIND_CONFIG ?= hack/kind-config.yaml 
KUSTOMIZE_VERSION ?= v4.5.7
# using v0.11.0 anything later breaks CRD validation - see https://github.com/kubernetes-sigs/controller-tools/pull/755
CONTROLLER_TOOLS_VERSION ?= v0.11.0
KIND_VERSION ?=v0.14.0
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"

OS := $(shell uname)
ifeq ($(OS),Darwin)
OS=darwin
else
OS=linux
endif

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: quickstart
quickstart: deps manifests generate build release docker-build cluster docker-load deploy ## Create cluster and all components with default values for quick dev environment

.PHONY: quickstart-local
quickstart-local: deps manifests generate build release docker-build cluster-local docker-load-local deploy ## *Only useful for offline development - same as quickstart but loads existing images locally 

.PHONY: example
example: release ## Run install example for users - follows docs
	$(KIND) create cluster --name=$(KIND_CLUSTER_NAME)
	kubectl wait --for=condition=ready pods --all -n kube-system
	$(KIND) load docker-image ${IMG} --name=$(KIND_CLUSTER_NAME)
	kubectl create -f dist/kai-deploy.yaml
	kubectl wait --for=condition=ready pods --all -n kai-system
	kubectl apply -f examples/http-echo/

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: go-deps
go-deps: ## Install go mod deps
	go get -v -t -d ./...

.PHONY: deps
deps: kind kustomize controller-gen envtest

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

.PHONY: cluster
cluster: kind ## Create kind cluster
	$(KIND) create cluster --config=$(KIND_CONFIG)

.PHONY: cluster-local
cluster-local: kind ## Create kind cluster
	$(KIND) create cluster --config=$(KIND_CONFIG)

.PHONY: clean
clean: kind ## deletes kind cluster
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)

.PHONY: forward
forward: ## port-forward localhost traffic to ingress
	kubectl -n $(GATEWAY_NAMESPACE) port-forward service/envoy-$(GATEWAY_NAME) 8888:8080

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

.PHONY: docker-build
docker-build: test ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

.PHONY: docker-load
docker-load: kind ## load docker image to kind cluster
	$(KIND) load docker-image ${IMG} --name=$(KIND_CLUSTER_NAME)

.PHONY: docker-load-local
docker-load-local: kind ## load docker image to kind cluster
	$(KIND) load docker-image ${IMG} --name=$(KIND_CLUSTER_NAME)
	$(KIND) load docker-image gcr.io/k8s-staging-gateway-api/admission-server:v0.6.0 --name=$(KIND_CLUSTER_NAME)
	$(KIND) load docker-image gcr.io/kubebuilder/kube-rbac-proxy:v0.13.1 --name=$(KIND_CLUSTER_NAME)
	$(KIND) load docker-image registry.k8s.io/ingress-nginx/kube-webhook-certgen:v1.1.1 --name=$(KIND_CLUSTER_NAME)
	$(KIND) load docker-image ghcr.io/projectcontour/contour:v1.24.3 --name=$(KIND_CLUSTER_NAME)
	$(KIND) load docker-image docker.io/envoyproxy/envoy:v1.25.4 --name=$(KIND_CLUSTER_NAME)
	$(KIND) load docker-image hashicorp/http-echo --name=$(KIND_CLUSTER_NAME)

# FIXME?: Should we check if buildx is installed first?
.PHONY: docker-buildx
docker-buildx: test ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- docker buildx create --name project-v3-builder
	docker buildx use project-v3-builder
	- docker buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- docker buildx rm project-v3-builder
	rm Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl create -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl create -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: release
release: manifests kustomize ## release generates single file yamls for crd's, controllers, and other resources for installing and deploying the operator
	mkdir -p dist
	$(KUSTOMIZE) build config/crd -o dist/crds.yaml
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default -o dist/kai-deploy.yaml

##@ Build Dependencies

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: kind
kind: $(KIND) ## Download kind locally if necessary
$(KIND): $(LOCALBIN)
	test -s $(LOCALBIN)/kind || curl -sSLo $(LOCALBIN)/kind https://kind.sigs.k8s.io/dl/$(KIND_VERSION)/kind-$(OS)-amd64 && chmod +x $(LOCALBIN)/kind
