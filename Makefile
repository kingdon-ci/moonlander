.PHONY: push all build fmt ko-build-push ko-build-tag ko-build-test
.PHONY: set-image build-and-set deploy deploy-secret ci-full ci

export KO_DOCKER_REPO := zot.urmanac.com:5050/moon/lander

build: fmt
	go build  cmd

fmt:
	gofmt -w  cmd/

push: all deploy

all: build-and-set

ko-build-push:
	ko build ./cmd/moonlander --bare

ko-build-tag:
	ko build ./cmd/moonlander --bare -t "$(TAG)"

ko-build-test:
	ko build ./cmd/moonlander --bare --push=false

set-image:
	@if [ -z "$(IMAGE)" ]; then \
		echo "ERROR: IMAGE variable is not set. Use 'make set-image IMAGE=<image>'"; \
		exit 1; \
	fi
	yq e -i '.spec.template.spec.containers[0].image = "$(IMAGE)"' deploy/basic/deployment.yaml

build-and-set:
	$(eval BUILT_IMAGE := $(shell make ko-build-push | tail -1))
	@echo "Built image: $(BUILT_IMAGE)"
	$(MAKE) set-image IMAGE=$(BUILT_IMAGE)

deploy:
	kubectl apply -k deploy/basic/

deploy-secret:
	kubectl apply -f deploy/basic/secret-regcred.yaml

ci-full: ci
	kind create cluster

ci:
	./hack/kind-build.sh
	./hack/kind-deploy.sh
	./hack/kind-moonlander-test-mocks.sh
	./hack/kind-test.sh
