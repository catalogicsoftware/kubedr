DOCKER_REGISTRY ?= docker-registry.devad.catalogic.us:5000

DOCKER_DIR_BASE = kubedr

DOCKER_KUBEDR_IMAGE_TAG ?= dev
DOCKER_KUBEDR_IMAGE_NAME_SHORT = kubedr
DOCKER_KUBEDR_IMAGE_NAME_LONG = ${DOCKER_REGISTRY}/${DOCKER_KUBEDR_IMAGE_NAME_SHORT}

DOCKER_KUBEDRUTIL_IMAGE_TAG ?= 0.0.2
DOCKER_KUBEDRUTIL_IMAGE_NAME_SHORT = kubedrutil
DOCKER_KUBEDRUTIL_IMAGE_NAME_LONG = ${DOCKER_REGISTRY}/${DOCKER_KUBEDRUTIL_IMAGE_NAME_SHORT}

build: docker_build go_build

docker_build:
	cd ${DOCKER_DIR_BASE} && \
		docker build \
			--tag ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:${DOCKER_KUBEDR_IMAGE_TAG} \
			.

docker_push_latest:
	docker pull ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:${DOCKER_KUBEDR_IMAGE_TAG} || true
	docker tag ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:${DOCKER_KUBEDR_IMAGE_TAG} ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:latest
	docker push ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:latest

docker_push_tags:
ifndef CI_COMMIT_TAG
	$(error The git tag, CI_COMMIT_TAG, is MISSING. This is required for pushing tagged images. Aborting.)
endif
	docker pull ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:${DOCKER_KUBEDR_IMAGE_TAG} || true
	docker tag ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:${DOCKER_KUBEDR_IMAGE_TAG} ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:${CI_COMMIT_TAG}
	docker push ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:${CI_COMMIT_TAG}

go_build:
	cd kubedr/config/manager && \
		kustomize edit set image controller=${DOCKER_KUBEDR_IMAGE_NAME_LONG}:${DOCKER_KUBEDR_IMAGE_TAG}
	cd kubedr && kustomize build config/default > kubedr.yaml
	sed -i 's#<KUBEDR_UTIL_IMAGE_VAL>#${DOCKER_KUBEDRUTIL_IMAGE_NAME_LONG}:${DOCKER_KUBEDRUTIL_IMAGE_TAG}#' kubedr/kubedr.yaml
