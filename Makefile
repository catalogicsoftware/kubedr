DOCKER_REGISTRY ?= docker-registry.devad.catalogic.us:5000

DOCKER_DIR_BASE = kubedr

DOCKER_KUBEDR_IMAGE_TAG ?= latest
DOCKER_KUBEDR_IMAGE_NAME_SHORT = kubedr
DOCKER_KUBEDR_IMAGE_NAME_LONG = ${DOCKER_REGISTRY}/${DOCKER_KUBEDR_IMAGE_NAME_SHORT}

DOCKER_KUBEDRUTIL_IMAGE_TAG ?= 0.0.1
DOCKER_KUBEDRUTIL_IMAGE_NAME_SHORT = kubedrutil
DOCKER_KUBEDRUTIL_IMAGE_NAME_LONG = ${DOCKER_REGISTRY}/${DOCKER_KUBEDRUTIL_IMAGE_NAME_SHORT}

# make >= 3.8.2
# Add special target to have make invoke one instance of shell, regardless of lines
.ONESHELL:
docker_build:
	cd ${DOCKER_DIR_BASE}
	docker pull ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:latest || true
	docker build \
		--cache-from ${DOCKER_KUBEDR_IMAGE_NAME_LONG}:latest \
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
	sed -i 's#DOCKER_KUBEDRUTIL_IMAGE_TAG#${DOCKER_KUBEDRUTIL_IMAGE_TAG}#' kubedr/config/manager/manager.yaml
	sed -i 's#DOCKER_KUBEDRUTIL_IMAGE_NAME_SHORT#${DOCKER_KUBEDRUTIL_IMAGE_NAME_SHORT}#' kubedr/config/manager/manager.yaml
	sed -i 's#DOCKER_KUBEDRUTIL_IMAGE_NAME_LONG#${DOCKER_KUBEDRUTIL_IMAGE_NAME_LONG}#' kubedr/config/manager/manager.yaml

	cd kubedr/config/manager
	kustomize edit set image controller=${DOCKER_KUBEDR_IMAGE_NAME_LONG}:${DOCKER_KUBEDR_IMAGE_TAG}
	cd ../../
	kustomize build config/default > kubedr.yaml
