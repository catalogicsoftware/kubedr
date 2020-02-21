DOCKER_PREFIX ?= catalogicsoftware/

DOCKER_DIR_BASE ?= kubedr

DOCKER_KUBEDR_IMAGE_TAG ?= latest
DOCKER_KUBEDR_IMAGE_NAME_SHORT ?= kubedr
DOCKER_KUBEDR_IMAGE_NAME_LONG ?= ${DOCKER_PREFIX}${DOCKER_KUBEDR_IMAGE_NAME_SHORT}

DOCKER_KUBEDRUTIL_IMAGE_TAG ?= 0.2.11
DOCKER_KUBEDRUTIL_IMAGE_NAME_SHORT ?= kubedrutil
DOCKER_KUBEDRUTIL_IMAGE_NAME_LONG ?= ${DOCKER_PREFIX}${DOCKER_KUBEDRUTIL_IMAGE_NAME_SHORT}

build: manifests docker_build go_build

manifests:
	cd ${DOCKER_DIR_BASE} && make manifests

docker_build:
	cd ${DOCKER_DIR_BASE} && \
		docker build \
			--tag ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:${DOCKER_KUBEDR_IMAGE_TAG} \
			.

go_build:
	cd kubedr/config/manager && \
		kustomize edit set image controller=${DOCKER_KUBEDR_IMAGE_NAME_LONG}:${DOCKER_KUBEDR_IMAGE_TAG}
	cd kubedr && kustomize build config/default > kubedr.yaml
	sed -i 's#<KUBEDR_UTIL_IMAGE_VAL>#${DOCKER_KUBEDRUTIL_IMAGE_NAME_LONG}:${DOCKER_KUBEDRUTIL_IMAGE_TAG}#' kubedr/kubedr.yaml
