IMG ?= observatory-operator:latest
LOCALBIN ?= $(shell pwd)/bin

SHELL := /usr/bin/env bash -o pipefail
.SHELLFLAGS := -ec

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

.PHONY: controller-gen kustomize envtest golangci-lint ## Install toolchain
controller-gen:
	@test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0
kustomize:
	@test -s $(LOCALBIN)/kustomize || { curl -sS https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh | bash -s -- 5.2.1 $(LOCALBIN); }
envtest:
	@test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
golangci-lint:
	@test -s $(LOCALBIN)/golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCALBIN) v1.55.2

.PHONY: manifests ## Generate CRDs
manifests: controller-gen
	$(LOCALBIN)/controller-gen rbac:roleName=observatory-controller crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: build ## Build manager
build:
	go build -o bin/manager cmd/main.go

.PHONY: run ## Run locally
run:
	go run ./cmd/main.go

.PHONY: docker-build ## Build image
docker-build:
	docker build -t $(IMG) .

.PHONY: docker-push ## Push image
docker-push:
	docker push $(IMG)

.PHONY: install ## Install CRDs
install: kustomize manifests
	$(LOCALBIN)/kustomize build config/crd | kubectl apply -f -

.PHONY: uninstall ## Uninstall CRDs
uninstall: kustomize
	$(LOCALBIN)/kustomize build config/crd | kubectl delete -f - --ignore-not-found

.PHONY: deploy ## Deploy controller
deploy: kustomize
	cd config/manager && $(LOCALBIN)/kustomize edit set image controller=$(IMG)
	$(LOCALBIN)/kustomize build config/default | kubectl apply -f -

.PHONY: undeploy ## Remove controller
undeploy: kustomize
	$(LOCALBIN)/kustomize build config/default | kubectl delete -f - --ignore-not-found

.PHONY: kind-create ## Create dev cluster
kind-create:
	kind create cluster --name observatory-dev

.PHONY: run-example-simple ## Run sample
run-example-simple:
	kubectl apply -f config/samples/simple-workflow.yaml

.PHONY: test validate-good validate-bad

test: ## Run unit tests for validator
	go test ./internal/validation -v

validate-good: ## Validate a known-good sample
	go run ./cmd/validate --root . ./sampledata/sample_event.json

validate-bad: ## Validate a known-bad sample (intentionally fails, but make continues)
	- go run ./cmd/validate --root . ./sampledata/sample_event_invalid.json || true

	