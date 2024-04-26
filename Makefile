## template Makefile:
## service example
#:

SHELL          = /bin/sh
CFG           ?= .env
PRG           ?= $(shell basename $$PWD)

# ------------------------------------------------------------------------------
# Include go source commands
SOURCES_EXISTS ?= $(shell find -name go.mod)

ifneq (,$(SOURCES_EXISTS))
include Makefile.golang
endif


# ------------------------------------------------------------------------------
## Docker build operations
#:

# build docker image directly
docker: $(PRG)
	docker build -t $(PRG) .

ALLARCH_DOCKER ?= "linux/amd64,linux/arm/v7,linux/arm64"

# build multiarch docker images via buildx
docker-multi:
	time docker buildx build --platform $(ALLARCH_DOCKER) -t $(DOCKER_IMAGE):$(APP_VERSION) --push .

# ------------------------------------------------------------------------------
## Other
#:

## update docs at pkg.go.dev
godoc:
	vf=$(APP_VERSION) ; v=$${vf%%-*} ; echo "Update for $$v..." ; \
	curl 'https://proxy.golang.org/$(GODOC_REPO)/@v/'$$v'.info'

## update latest docker image tag at ghcr.io
ghcr:
	v=$(APP_VERSION) ; echo "Update for $$v..." ; \
	docker pull $(DOCKER_IMAGE):$$v && \
	docker tag $(DOCKER_IMAGE):$$v $(DOCKER_IMAGE):latest && \
	docker push $(DOCKER_IMAGE):latest

# ------------------------------------------------------------------------------

# Load AUTH_TOKEN
-include $(DCAPE_ROOT)/var/oauth2-token

# create OAuth application credentials
oauth2-create:
	$(MAKE) -s oauth2-app-create HOST=$(AS_HOST) URL=/login PREFIX=AS
