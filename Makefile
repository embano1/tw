# Inspired by https://github.com/jessfraz/reg/blob/master/Makefile
NAME := tw
PKG := github.com/embano1/$(NAME)
DOCKER_IMAGE := embano1/tw

VERSION := $(shell cat VERSION)
GITCOMMIT := $(shell git rev-parse --short HEAD)
BUILDDATE := $(shell date +%D-%H:%M)

# Set any default go build tags
BUILDTAGS :=

CTIMEVAR=-X main.commit=$(GITCOMMIT) -X main.version=$(VERSION) -X main.date=$(BUILDDATE)
GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

# Default make target
default: static

## Builds a dynamic executable or package
.PHONY: build
build: $(NAME) 
$(NAME): *.go VERSION
	@echo "+ $@"
	go build -tags "$(BUILDTAGS)" ${GO_LDFLAGS} -o $(NAME) .

## Builds a static executable
.PHONY: static
static: 
	@echo "+ $@"
	CGO_ENABLED=0 go build \
				-tags "$(BUILDTAGS) static_build" \
				${GO_LDFLAGS_STATIC} -o $(NAME) .

## Cleanup any build binaries or packages
.PHONY: clean
clean: 
	@echo "+ $@"
	$(RM) $(NAME)

## Check for a clean git repository state
.PHONY: git-no-dirty
git-no-dirty:
	$(if $(shell git status --porcelain), $(error "GIT is still dirty. Please clean before call make."), @true)

## Build Docker image
.PHONY: image
image: git-no-dirty
	docker build --rm --force-rm -f Dockerfile -t $(DOCKER_IMAGE):$(GITCOMMIT) .
	docker tag $(DOCKER_IMAGE):$(GITCOMMIT) $(DOCKER_IMAGE):latest

## Push Docker image
.PHONY: push-image
push-image: image
	docker push $(DOCKER_IMAGE):latest
	docker push $(DOCKER_IMAGE):$(GITCOMMIT)
