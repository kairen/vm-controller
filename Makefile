VERSION_MAJOR ?= 0
VERSION_MINOR ?= 1
VERSION_BUILD ?= 0
VERSION ?= v$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_BUILD)

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

ORG := github.com
OWNER := kairen
REPOPATH ?= $(ORG)/$(OWNER)/controller

$(shell mkdir -p ./out)

.PHONY: dep 
dep:
	dep ensure -v

.PHONY: build
build: out/controller out/apiserver

.PHONY: out/controller
out/controller:
	CGO_ENABLED=0 GOOS=$* GOARCH=$(GOARCH) go build \
	  -ldflags="-s -w -X $(REPOPATH)/pkg/version.version=$(VERSION)" \
	  -a -o $@ cmd/controller/main.go

.PHONY: out/apiserver
out/apiserver:
	CGO_ENABLED=0 GOOS=$* GOARCH=$(GOARCH) go build \
	  -ldflags="-s -w -X $(REPOPATH)/pkg/version.version=$(VERSION)" \
	  -a -o $@ cmd/apiserver/main.go

.PHONY: build_images
build_images: ctrl_image apiserver_image

.PHONY: push_image
push_images:
	docker push kairen/vm-controller:$(VERSION)
	docker push kairen/vm-apiserver:$(VERSION)

.PHONY: ctrl_image
ctrl_image:
	docker build -t kairen/vm-controller:$(VERSION) .

.PHONY: apiserver_image
apiserver_image:
	docker build --file Dockerfile.apiserver -t kairen/vm-apiserver:$(VERSION) .

.PHONY: clean
clean:
	rm -rf out/

