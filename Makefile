
BUILD_VERSION ?= latest
DOCKER_REGISTRY := docker-registry.devad.catalogic.us:5000
KUBEDR_IMAGE_NAME := kubedr
KUBEDRUTIL_IMAGE_NAME := kubedrutil

KUBEDR_FULL_IMAGE_NAME := $(DOCKER_REGISTRY)/$(KUBEDR_IMAGE_NAME):$(BUILD_VERSION)
KUBEDRUTIL_FULL_IMAGE_NAME := $(DOCKER_REGISTRY)/$(KUBEDRUTIL_IMAGE_NAME):$(BUILD_VERSION)

build:
	(cd containers/kubedrutil; docker build -t $(KUBEDRUTIL_FULL_IMAGE_NAME) .)
	sed -i 's#<KUBEDR_UTIL_IMAGE>#$(KUBEDRUTIL_FULL_IMAGE_NAME)#' kubedr/config/manager/manager.yaml
	(cd kubedr; make build IMG=$(KUBEDR_FULL_IMAGE_NAME))

pushimage:
	docker push $(KUBEDR_FULL_IMAGE_NAME)
	docker push $(KUBEDRUTIL_FULL_IMAGE_NAME)
